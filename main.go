package main

import (
	"encoding/json"
	//	"fmt"
	"flag"
	"github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"log"
	"net/http"
	"time"
)

var (
	signKey = flag.String("signKey", "lhpp", "token signKey")
)

func main() {

	StartServer()

}

func StartServer() {
	flag.Parse()
	//新建路由
	r := mux.NewRouter()

	//新建jwt中间件
	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return []byte(*signKey), nil
		},
		SigningMethod: jwt.SigningMethodHS256,
		//http header authorization bearer 的方式
		Extractor: jwtmiddleware.FromFirst(jwtmiddleware.FromAuthHeader,
			jwtmiddleware.FromParameter("auth_code")),
	})
	//n := negroni.Classic()
	//e := casbin.NewEnforcer("authz_model.conf", "authz_policy.csv")

	r.HandleFunc("/ping", PingHandler)
	//j := negroni.HandlerFunc(jwtMiddleware.HandlerWithNext)
	r.Handle("/secured/ping", negroni.New(
		negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
		//negroni.HandlerFunc(authz.Authorizer(e)),
		negroni.Wrap(http.HandlerFunc(SecuredPingHandler)),
	))
	//n.Use(j)
	//n.UseHandler(r)
	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":6600", nil))
}

type Response struct {
	Text string `json:"text"`
}

func respondJson(text string, w http.ResponseWriter) {
	response := Response{text}

	jsonResponse, err := json.Marshal(response)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

func PingHandler(w http.ResponseWriter, r *http.Request) {
	mySigningKey := []byte(*signKey)
	type MyCustomClaims struct {
		Foo string `json:"foo"`
		jwt.StandardClaims
	}

	// Create the Claims
	claims := MyCustomClaims{
		"bar",
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Second * 10).Unix(),
			Issuer:    "test",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString(mySigningKey)
	if err != nil {
		panic(nil)
	}
	respondJson(ss, w)

}

func SecuredPingHandler(w http.ResponseWriter, r *http.Request) {
	respondJson("All good. You only get this message if you're authenticated", w)
}
