package auth

import (
	"fmt"
	"net/http"
	"time"
	
	"github.com/dchest/uniuri"
	"github.com/matthewhartstonge/argon2"
	
	"github.com/daarlabs/hirokit/cache"
	"github.com/daarlabs/hirokit/cookie"
	"github.com/daarlabs/hirokit/esquel"
)

type Manager interface {
	Session() SessionManager
	Tfa() TfaManager
	User() UserManager
	CustomUser(id int, email string) UserManager
	Manager() UserManager
	
	In(email, password string, roles ...Role) (In, error)
	Out() error
	
	MustIn(email, password string, roles ...Role) In
	MustOut()
}

type In struct {
	Token string
	Ok    bool
	Tfa   bool `json:"tfa"`
}

type manager struct {
	db     *esquel.DB
	req    *http.Request
	res    http.ResponseWriter
	cookie cookie.Cookie
	cache  cache.Client
	config Config
}

func New(
	db *esquel.DB,
	req *http.Request,
	res http.ResponseWriter,
	cookie cookie.Cookie,
	cache cache.Client,
	config Config,
) Manager {
	return &manager{
		db:     db,
		req:    req,
		res:    res,
		cookie: cookie,
		cache:  cache,
		config: config,
	}
}

func (m *manager) In(email, password string, roles ...Role) (In, error) {
	var r User
	err := esquel.New(m.db).
		Q(fmt.Sprintf(`SELECT id, email, roles, password, tfa FROM %s`, usersTable)).
		Q("WHERE email = @email", esquel.Map{"email": email}).
		Q("AND active = true").
		Exec(&r)
	if err != nil {
		return In{
			Ok:  false,
			Tfa: false,
		}, err
	}
	if r.Id == 0 {
		return In{
			Ok:  false,
			Tfa: false,
		}, ErrorInvalidCredentials
	}
	if len(roles) > 0 && len(r.Roles) > 0 {
		var exists bool
		for _, userRole := range r.Roles {
			for _, role := range roles {
				if userRole == role.Name {
					exists = true
					break
				}
			}
		}
		if !exists {
			return In{
				Ok:  false,
				Tfa: false,
			}, ErrorInvalidCredentials
		}
	}
	ok, err := argon2.VerifyEncoded([]byte(password), []byte(r.Password))
	if !ok || err != nil {
		return In{
			Ok:  false,
			Tfa: false,
		}, err
	}
	if r.Tfa {
		token := uniuri.New()
		if err := m.cache.Set(createTfaCacheKey(token), User{Id: r.Id}, time.Minute*5); err != nil {
			return In{}, err
		}
		m.cookie.Set(TfaCookieKey, token, time.Minute*5)
		return In{
			Token: token,
			Ok:    true,
			Tfa:   true,
		}, nil
	}
	token, err := m.Session().New(r)
	if err != nil {
		return In{}, err
	}
	if err := m.User().UpdateActivity(r.Id); err != nil {
		return In{}, err
	}
	return In{
		Token: token,
		Ok:    true,
		Tfa:   false,
	}, nil
}

func (m *manager) MustIn(email, password string, roles ...Role) In {
	r, err := m.In(email, password, roles...)
	if err != nil {
		panic(err)
	}
	return r
}

func (m *manager) Out() error {
	return m.Session().Destroy()
}

func (m *manager) MustOut() {
	err := m.Out()
	if err != nil {
		panic(err)
	}
}

func (m *manager) Session() SessionManager {
	return createSessionManager(
		m.req,
		m.res,
		m.cookie,
		m.cache,
		m.config,
	)
}

func (m *manager) Tfa() TfaManager {
	return createTfaManager(m)
}

func (m *manager) User() UserManager {
	session := m.Session().MustGet()
	return CreateUserManager(m.db, m.cache, session.Id, session.Email)
}

func (m *manager) CustomUser(id int, email string) UserManager {
	return CreateUserManager(m.db, m.cache, id, email)
}

func (m *manager) Manager() UserManager {
	return m.CustomUser(0, "")
}
