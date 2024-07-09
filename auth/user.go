package auth

import (
	"database/sql"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"time"
	
	"github.com/dchest/uniuri"
	"github.com/matthewhartstonge/argon2"
	
	"github.com/daarlabs/hirokit/cache"
	"github.com/daarlabs/hirokit/esquel"
)

type UserManager interface {
	Exists(id ...int) (bool, error)
	Get(id ...int) (User, error)
	Create(r User) (int, error)
	Update(r User, columns ...string) error
	ResetPassword(token ...string) (string, error)
	DestroyResetPassword(token string) error
	UpdatePassword(actualPassword, newPassword string) error
	ForceUpdatePassword(newPassword string) error
	Enable(id ...int) error
	Disable(id ...int) error
	UpdateActivity(id ...int) error
	
	MustExists(id ...int) bool
	MustGet(id ...int) User
	MustCreate(r User) int
	MustUpdate(r User, columns ...string)
	MustResetPassword(token ...string) string
	MustDestroyResetPassword(token string)
	MustUpdatePassword(actualPassword, newPassword string)
	MustForceUpdatePassword(newPassword string)
	MustEnable(id ...int)
	MustDisable(id ...int)
	MustUpdateActivity(id ...int)
}

type User struct {
	Id           int              `json:"id" db:"id"`
	Active       bool             `json:"active" db:"active"`
	Roles        []string         `json:"roles" db:"roles"`
	Email        string           `json:"email" db:"email"`
	Password     string           `json:"password" db:"password"`
	Tfa          bool             `json:"tfa" db:"tfa"`
	TfaSecret    sql.Null[string] `json:"tfaSecret" db:"tfa_secret"`
	TfaCodes     sql.Null[string] `json:"tfaCodes" db:"tfa_codes"`
	TfaUrl       sql.Null[string] `json:"tfaUrl" db:"tfa_url"`
	LastActivity time.Time        `json:"lastActivity" db:"last_activity"`
	CreatedAt    time.Time        `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time        `json:"updatedAt" db:"updated_at"`
}

type userManager struct {
	db         *esquel.DB
	cache      cache.Client
	id         int
	email      string
	driverName string
	data       esquel.Map
}

const (
	UserActive       = "active"
	UserRoles        = "roles"
	UserEmail        = "email"
	UserPassword     = "password"
	UserTfa          = "tfa"
	UserTfaSecret    = "tfa_secret"
	UserTfaCodes     = "tfa_codes"
	UserTfaUrl       = "tfa_url"
	UserLastActivity = "last_activity"
)

const (
	usersTable  = "users"
	paramPrefix = "@"
)

const (
	operationInsert = "insert"
	operationUpdate = "update"
)

var (
	argon = argon2.DefaultConfig()
)

func CreateUserManager(db *esquel.DB, cache cache.Client, id int, email string) UserManager {
	return &userManager{
		db:         db,
		cache:      cache,
		email:      email,
		id:         id,
		data:       make(map[string]any),
		driverName: db.DriverName(),
	}
}

func (u *userManager) Exists(id ...int) (bool, error) {
	user, err := u.Get(id...)
	if err != nil {
		return false, err
	}
	return user.Id > 0, nil
}

func (u *userManager) MustExists(id ...int) bool {
	exists, err := u.Exists(id...)
	if err != nil {
		panic(err)
	}
	return exists
}

func (u *userManager) Get(id ...int) (User, error) {
	if len(id) > 0 {
		u.id = id[0]
	}
	var r User
	if u.id == 0 && u.email == "" {
		return r, ErrorInvalidUser
	}
	err := esquel.New(u.db).Q(`SELECT *`).
		Q(fmt.Sprintf(`FROM %s`, usersTable)).
		If(u.id > 0, `WHERE id = @id`, esquel.Map{"id": u.id}).
		If(u.id == 0, `WHERE email = @email`, esquel.Map{"email": u.email}).
		Q(`LIMIT 1`).
		Exec(&r)
	clear(u.data)
	return r, err
}

func (u *userManager) MustGet(id ...int) User {
	r, err := u.Get(id...)
	if err != nil {
		panic(err)
	}
	return r
}

func (u *userManager) Create(r User) (int, error) {
	if u.id != 0 {
		return u.id, ErrorUserAlreadyExists
	}
	if err := u.readData(operationInsert, r, []string{}); err != nil {
		return 0, err
	}
	columns, placeholders := u.insertValues()
	err := esquel.New(u.db).Q(fmt.Sprintf(`INSERT INTO %s`, usersTable)).
		Q(fmt.Sprintf(`(%s)`, columns)).
		Q(fmt.Sprintf(`VALUES (%s)`, placeholders), u.args()...).
		Q(`RETURNING id`).
		Exec(&u.id)
	u.email = r.Email
	clear(u.data)
	return u.id, err
}

func (u *userManager) MustCreate(r User) int {
	id, err := u.Create(r)
	if err != nil {
		panic(err)
	}
	return id
}

func (u *userManager) Update(r User, columns ...string) error {
	if u.id == 0 && u.email == "" {
		return ErrorInvalidUser
	}
	if err := u.readData(operationUpdate, r, columns); err != nil {
		return err
	}
	err := esquel.New(u.db).Q(fmt.Sprintf(`UPDATE %s`, usersTable)).
		Q(fmt.Sprintf(`SET %s`, u.updateValues()), u.args()...).
		If(u.id > 0, `WHERE id = @id`, esquel.Map{"id": u.id}).
		If(u.id == 0, `WHERE email = @email`, esquel.Map{"email": u.email}).
		Exec()
	clear(u.data)
	return err
}

func (u *userManager) MustUpdate(r User, columns ...string) {
	err := u.Update(r, columns...)
	if err != nil {
		panic(err)
	}
}

func (u *userManager) ResetPassword(token ...string) (string, error) {
	if len(token) > 0 {
		var r User
		err := u.cache.Get(u.createResetPasswordKey(token[0]), &r)
		u.email = r.Email
		return r.Email, err
	}
	t := uniuri.New()
	return t, u.cache.Set(
		u.createResetPasswordKey(t),
		User{Email: u.email},
		time.Hour,
	)
}

func (u *userManager) MustResetPassword(token ...string) string {
	t, err := u.ResetPassword(token...)
	if err != nil {
		panic(err)
	}
	return t
}

func (u *userManager) DestroyResetPassword(token string) error {
	return u.cache.Destroy(u.createResetPasswordKey(token))
}

func (u *userManager) MustDestroyResetPassword(token string) {
	err := u.DestroyResetPassword(token)
	if err != nil {
		panic(err)
	}
}

func (u *userManager) UpdatePassword(actualPassword, newPassword string) error {
	if u.id == 0 && u.email == "" {
		return ErrorMissingUser
	}
	user, err := u.Get()
	if err != nil {
		return err
	}
	if ok, err := argon2.VerifyEncoded([]byte(actualPassword), []byte(user.Password)); !ok || err != nil {
		return ErrorMismatchPassword
	}
	hash, err := u.hashPassword(newPassword)
	if err != nil {
		return err
	}
	err = esquel.New(u.db).Q(fmt.Sprintf(`UPDATE %s`, usersTable)).
		Q(`SET password = @password`, esquel.Map{"password": hash}).
		If(u.id > 0, `WHERE id = @id`, esquel.Map{"id": u.id}).
		If(u.id == 0, `WHERE email = @email`, esquel.Map{"email": u.email}).
		Exec()
	clear(u.data)
	return err
}

func (u *userManager) MustUpdatePassword(actualPassword, newPassword string) {
	err := u.UpdatePassword(actualPassword, newPassword)
	if err != nil {
		panic(err)
	}
}

func (u *userManager) ForceUpdatePassword(newPassword string) error {
	if u.id == 0 && u.email == "" {
		return ErrorMissingUser
	}
	hash, err := u.hashPassword(newPassword)
	if err != nil {
		return err
	}
	err = esquel.New(u.db).Q(fmt.Sprintf(`UPDATE %s`, usersTable)).
		Q(`SET password = @password`, esquel.Map{"password": hash}).
		If(u.id > 0, `WHERE id = @id`, esquel.Map{"id": u.id}).
		If(u.id == 0, `WHERE email = @email`, esquel.Map{"email": u.email}).
		Exec()
	clear(u.data)
	return err
}

func (u *userManager) MustForceUpdatePassword(newPassword string) {
	err := u.ForceUpdatePassword(newPassword)
	if err != nil {
		panic(err)
	}
}

func (u *userManager) Enable(id ...int) error {
	if len(id) > 0 {
		u.id = id[0]
	}
	if u.id == 0 && u.email == "" {
		return ErrorInvalidUser
	}
	err := esquel.New(u.db).Q(fmt.Sprintf(`UPDATE %s`, usersTable)).
		Q(`SET active = true`).
		If(u.id > 0, `WHERE id = @id`, esquel.Map{"id": u.id}).
		If(u.id == 0, `WHERE email = @email`, esquel.Map{"email": u.email}).
		Exec()
	clear(u.data)
	return err
}

func (u *userManager) MustEnable(id ...int) {
	err := u.Enable(id...)
	if err != nil {
		panic(err)
	}
}

func (u *userManager) Disable(id ...int) error {
	if len(id) > 0 {
		u.id = id[0]
	}
	if u.id == 0 && u.email == "" {
		return ErrorInvalidUser
	}
	err := esquel.New(u.db).Q(fmt.Sprintf(`UPDATE %s`, usersTable)).
		Q(`SET active = false`).
		If(u.id > 0, `WHERE id = @id`, esquel.Map{"id": u.id}).
		If(u.id == 0, `WHERE email = @email`, esquel.Map{"email": u.email}).
		Exec()
	clear(u.data)
	return err
}

func (u *userManager) MustDisable(id ...int) {
	err := u.Disable(id...)
	if err != nil {
		panic(err)
	}
}

func (u *userManager) UpdateActivity(id ...int) error {
	if len(id) > 0 {
		u.id = id[0]
	}
	if u.id == 0 && u.email == "" {
		return ErrorInvalidUser
	}
	err := esquel.New(u.db).Q(fmt.Sprintf(`UPDATE %s`, usersTable)).
		Q(`SET last_activity = CURRENT_TIMESTAMP`).
		If(u.id > 0, `WHERE id = @id`, esquel.Map{"id": u.id}).
		If(u.id == 0, `WHERE email = @email`, esquel.Map{"email": u.email}).
		Exec()
	clear(u.data)
	return err
}

func (u *userManager) MustUpdateActivity(id ...int) {
	err := u.UpdateActivity(id...)
	if err != nil {
		panic(err)
	}
}

func (u *userManager) readData(operation string, data User, columns []string) error {
	columnsExist := len(columns) > 0
	if operation == operationInsert && slices.Contains(columns, esquel.Id) {
		u.data[esquel.Id] = data.Id
	}
	if !columnsExist || slices.Contains(columns, UserActive) {
		u.data[UserActive] = data.Active
	}
	if !columnsExist || slices.Contains(columns, UserEmail) {
		u.data[UserEmail] = data.Email
	}
	if !columnsExist || slices.Contains(columns, UserPassword) {
		hash, err := u.hashPassword(data.Password)
		if err != nil {
			return err
		}
		u.data[UserPassword] = hash
	}
	if !columnsExist || slices.Contains(columns, UserRoles) {
		u.data[UserRoles] = data.Roles
	}
	if !columnsExist || slices.Contains(columns, UserTfa) {
		u.data[UserTfa] = data.Tfa
	}
	if !columnsExist || slices.Contains(columns, UserTfaSecret) {
		u.data[UserTfaSecret] = data.TfaSecret.V
	}
	if !columnsExist || slices.Contains(columns, UserTfaCodes) {
		u.data[UserTfaCodes] = data.TfaCodes.V
	}
	if !columnsExist || slices.Contains(columns, UserTfaUrl) {
		u.data[UserTfaUrl] = data.TfaUrl.V
	}
	return nil
}

func (u *userManager) hashPassword(password string) (string, error) {
	if strings.HasPrefix(password, "$argon2") {
		return password, nil
	}
	hash, err := argon.HashEncoded([]byte(password))
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func (u *userManager) args() []esquel.Map {
	if len(u.data) == 0 {
		return []esquel.Map{}
	}
	result := u.data
	vectors := make([]any, 0)
	for name, v := range u.data {
		kind := reflect.TypeOf(v).Kind()
		if name == UserPassword || kind == reflect.Bool {
			continue
		}
		vectors = append(vectors, v)
	}
	switch u.driverName {
	case esquel.Postgres:
		if len(vectors) > 0 {
			result[esquel.Vectors] = esquel.CreateTsVector(vectors...)
		}
	}
	return []esquel.Map{result}
}

func (u *userManager) insertValues() (string, string) {
	columns := []string{esquel.Id}
	placeholders := []string{esquel.Default}
	for name := range u.data {
		columns = append(columns, name)
		placeholders = append(placeholders, paramPrefix+name)
	}
	switch u.driverName {
	case esquel.Postgres:
		if len(u.data) > 0 {
			columns = append(columns, esquel.Vectors)
			placeholders = append(placeholders, fmt.Sprintf("to_tsvector(%s%s)", paramPrefix, esquel.Vectors))
		}
	}
	columns = append(columns, UserLastActivity)
	placeholders = append(placeholders, esquel.CurrentTimestamp)
	
	columns = append(columns, esquel.CreatedAt)
	placeholders = append(placeholders, esquel.CurrentTimestamp)
	
	columns = append(columns, esquel.UpdatedAt)
	placeholders = append(placeholders, esquel.CurrentTimestamp)
	return strings.Join(columns, ","), strings.Join(placeholders, ",")
}

func (u *userManager) updateValues() string {
	result := make([]string, 0)
	for column := range u.data {
		if column == esquel.Id {
			continue
		}
		result = append(result, fmt.Sprintf("%s = %s%s", column, paramPrefix, column))
	}
	result = append(result, fmt.Sprintf("%s = %s", UserLastActivity, esquel.CurrentTimestamp))
	result = append(result, fmt.Sprintf("%s = %s", esquel.UpdatedAt, esquel.CurrentTimestamp))
	switch u.driverName {
	case esquel.Postgres:
		vectors := make([]any, 0)
		for column, v := range u.data {
			if column == esquel.Id {
				continue
			}
			vectors = append(vectors, v)
		}
		if len(vectors) > 0 {
			result = append(result, fmt.Sprintf("%s = to_tsvector(%s%s)", esquel.Vectors, paramPrefix, esquel.Vectors))
		}
	}
	return strings.Join(result, ",")
}

func (u *userManager) createResetPasswordKey(token string) string {
	return "reset-password:" + token
}
