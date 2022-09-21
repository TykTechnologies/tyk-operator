package integration

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/cenkalti/backoff/v4"
	"sigs.k8s.io/e2e-framework/pkg/envconf"
)

const (
	proAdminSecret = "54321"
	ceAdminSecret  = "foo"
	orgOwner       = "Tyk operator e2e"
	portalCname    = "tyk-portal.local:7200/"
)

type Organization struct {
	ID           string `json:"id,omitempty"`
	Owner        string `json:"owner_name"`
	OwnerSlug    string `json:"owner_slug"`
	CnameEnabled bool   `json:"cname_enabled"`
	Cname        string `json:"cname"`
}

type User struct {
	ID           string `json:"id,omitempty"`
	FirstName    string `json:"first_name" `
	LastName     string `json:"last_name"`
	EmailAddress string `json:"email_address"`
	Password     string `json:"password"`
	OrgID        string `json:"org_id"`
	Active       bool   `json:"active"`
}

type E2EContext struct {
	Super struct {
		Access string
		User   *User
	}
	Orgs  map[string]*Organization
	Users map[string]*User
}
type E2EContextKey struct{}

func set(ctx context.Context, e *E2EContext) context.Context {
	return context.WithValue(ctx, E2EContextKey{}, e)
}

func get(ctx context.Context) *E2EContext {
	return ctx.Value(E2EContextKey{}).(*E2EContext)
}

func (e *E2EContext) CreateSuperUser() error {
	n := make([]byte, 4)

	_, err := rand.Read(n)
	if err != nil {
		return err
	}

	// create super user
	super := &User{
		FirstName:    "super",
		LastName:     "user",
		EmailAddress: fmt.Sprintf("%x@operator.test", n),
		Password:     "newpassword",
		OrgID:        "",
		Active:       true,
	}

	access, err := e.createUser(super)
	if err != nil {
		return err
	}

	e.Super.Access = access
	e.Super.User = super

	return nil
}

func (e *E2EContext) Setup() error {
	if isPro() {
		err := e.CreateSuperUser()
		if err != nil {
			return err
		}

		err = e.DeleteExistingOrgs()
		if err != nil {
			return err
		}

		err = e.DeleteExistingUsers()
		if err != nil {
			return err
		}

		if err := e.CreateOrganizations(); err != nil {
			return err
		}

		if err := e.CreateUsers(); err != nil {
			return err
		}
	}

	return nil
}

func (e *E2EContext) CreateOrganizations() error {
	e.Orgs = make(map[string]*Organization)

	for i := 0; i < 3; i++ {
		o := &Organization{
			Owner:        fmt.Sprintf("%s-%d", orgOwner, i),
			CnameEnabled: true,
			OwnerSlug:    fmt.Sprint(i),
			Cname:        fmt.Sprintf("%d%v", i, portalCname),
		}

		if err := e.createOrg(o); err != nil {
			return err
		}

		e.Orgs[o.ID] = o
	}

	return nil
}

func (e *E2EContext) CreateUsers() error {
	e.Users = make(map[string]*User)
	ek := "u"

	for k, v := range e.Orgs {
		ek += ek
		o := &User{
			FirstName:    v.Owner,
			LastName:     v.Owner,
			EmailAddress: ek + "@example.com",
			Password:     "newpassword",
			OrgID:        k,
			Active:       true,
		}

		access, err := e.createUser(o)
		if err != nil {
			return err
		}

		e.Users[access] = o
	}

	return nil
}

type ApiError struct {
	Status  string
	Message string
	Meta    interface{}
	Errors  []string
}

func (e *E2EContext) createOrg(o *Organization) (err error) {
	fmt.Print("Creating Organization "+o.Owner, "...")

	defer func() {
		if err != nil {
			fmt.Printf("ERROR %v\n", err)
		} else {
			fmt.Printf("Ok %v\n", o.ID)
		}
	}()

	var b bytes.Buffer

	json.NewEncoder(&b).Encode(o)

	var res *http.Response

	res, err = e.AdminPost("/admin/organisations", b.Bytes())
	if err != nil {
		return
	}

	defer res.Body.Close()

	var a ApiError
	json.NewDecoder(res.Body).Decode(&a)

	if strings.ToLower(a.Status) != "ok" {
		err = errors.New(a.Message)
		return
	}

	o.ID = a.Meta.(string)

	return nil
}

func (e *E2EContext) createUser(o *User) (access string, err error) {
	fmt.Print("Creating User "+o.FirstName, "...")

	defer func() {
		if err != nil {
			fmt.Printf("ERROR %v\n", err)
		} else {
			fmt.Printf("Ok %v\n", o.ID)
		}
	}()

	var b bytes.Buffer

	json.NewEncoder(&b).Encode(o)

	var res *http.Response

	res, err = e.AdminPost("/admin/users", b.Bytes())
	if err != nil {
		return
	}

	defer res.Body.Close()
	var a ApiError

	json.NewDecoder(res.Body).Decode(&a)

	if strings.ToLower(a.Status) != "ok" {
		return "", fmt.Errorf("%v", a)
	}

	o.ID = a.Meta.(map[string]interface{})["id"].(string)
	access = a.Message

	return
}

func (e E2EContext) DeleteExistingOrgs() error {
	ls := struct {
		Orgs []*Organization `json:"organisations"`
	}{}

	res, err := e.AdminGet("/admin/organisations", nil)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	json.NewDecoder(res.Body).Decode(&ls)

	for _, v := range ls.Orgs {
		if strings.Contains(v.Owner, orgOwner) {
			errDel := e.deleteOrg(v.Owner, v.ID)
			if errDel != nil {
				return errDel
			}
		}
	}

	return nil
}

func (e E2EContext) AdminGet(path string, body []byte) (*http.Response, error) {
	return e.adminReq(http.MethodGet, path, body)
}

func (e E2EContext) AdminDelete(path string, body []byte) (*http.Response, error) {
	return e.adminReq(http.MethodDelete, path, body)
}

func (e E2EContext) AdminPost(path string, body []byte) (*http.Response, error) {
	return e.adminReq(http.MethodPost, path, body)
}

func (e *E2EContext) adminReq(method, url string, body []byte) (res *http.Response, err error) {
	url = adminLocalhost + url
	err = backoff.Retry(func() error {
		var rd io.Reader
		if body != nil {
			rd = bytes.NewReader(body)
		}

		r, err := http.NewRequest(method, url, rd)
		if err != nil {
			return backoff.Permanent(err)
		}

		res, err = e.adminClient().Do(r)

		return err
	}, backoff.NewExponentialBackOff())

	return
}

var client = &http.Client{}

type e3eClient struct {
	modify func(r *http.Request)
}

func (e *e3eClient) Do(r *http.Request) (res *http.Response, err error) {
	e.modify(r)
	return client.Do(r)
}

func (e *E2EContext) DeleteExistingUsers() error {
	res, err := e.AdminGet("/api/users", nil)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	ls := struct {
		ApiError
		Users []*User `json:"users,omitempty"`
	}{}

	json.NewDecoder(res.Body).Decode(&ls)

	if ls.Message != "" {
		return fmt.Errorf("%v", ls)
	}

	for _, u := range ls.Users {
		if strings.Contains(u.FirstName, orgOwner) {
			if err := e.deleteUser(u); err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *E2EContext) adminClient() *e3eClient {
	if isPro() {
		return &e3eClient{modify: func(r *http.Request) {
			r.Header.Set("admin-auth", proAdminSecret)
			r.Header.Set("Authorization", e.Super.Access)
		}}
	}

	return &e3eClient{modify: func(r *http.Request) {
		r.Header.Set("X-Tyk-Authorization", ceAdminSecret)
	}}
}

func (e *E2EContext) deleteUser(o *User) (err error) {
	fmt.Print("Deleting User "+o.FirstName, "...")

	defer func() {
		if err != nil {
			fmt.Printf("ERROR %v\n", err)
		} else {
			fmt.Printf("Ok %v\n", o.ID)
		}
	}()

	var res *http.Response

	res, err = e.AdminDelete("/api/users/"+o.ID, nil)
	if err != nil {
		return
	}
	var a ApiError

	json.NewDecoder(res.Body).Decode(&a)
	res.Body.Close()

	if strings.ToLower(a.Status) != "ok" {
		err = fmt.Errorf("%v", err)
	}

	return
}

func (e *E2EContext) deleteOrg(name, id string) (err error) {
	fmt.Print("Deleting Organization "+name, "...")

	defer func() {
		if err != nil {
			fmt.Printf("ERROR %v\n", err)
		} else {
			fmt.Printf("Ok %v\n", id)
		}
	}()

	var res *http.Response

	res, err = e.AdminDelete("/admin/organisations/"+id, nil)
	if err != nil {
		return
	}

	var a ApiError

	json.NewDecoder(res.Body).Decode(&a)
	res.Body.Close()

	if strings.ToLower(a.Status) != "ok" {
		err = fmt.Errorf("%v", err)
	}

	return
}

func (e *E2EContext) Cleanup() error {
	if isPro() {
		for _, u := range e.Users {
			e.deleteUser(u)
		}

		for _, v := range e.Orgs {
			err := e.deleteOrg(v.Owner, v.ID)
			if err != nil {
				return err
			}
		}

		return e.deleteUser(e.Super.User)
	}

	return nil
}

func setupE2E(c1 context.Context, c2 *envconf.Config) (context.Context, error) {
	e := &E2EContext{}

	err := e.Setup()
	if err != nil {
		return c1, fmt.Errorf("Failed setting up e2e context %e", err)
	}

	return set(c1, e), nil
}

func teardownE2E(c1 context.Context, c2 *envconf.Config) (context.Context, error) {
	return c1, get(c1).Cleanup()
}
