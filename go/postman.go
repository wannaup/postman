package main

import (
    "os"
    "fmt"
    "html"
    "log"
    "net/http"
    "io"
    "encoding/json"
    "encoding/base64"
    "strings"
    "github.com/codegangsta/negroni"
    "github.com/gorilla/mux"
    "github.com/gorilla/context"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
)



type key int

const db key = 0
const userId key = 1
var config map[string]string

func main() {
    LoadConfig("conf_debug.json", &config)
    router := mux.NewRouter().StrictSlash(true)
    router.HandleFunc("/inbound", ProcessInbound).Methods("POST")
    router.HandleFunc("/threads", CreateThread).Methods("POST")
    router.HandleFunc("/threads", GetAllThreads).Methods("GET")
    router.HandleFunc("/threads/{threadId}", GetOneThread).Methods("GET")
    router.HandleFunc("/threads/{threadId}/reply", ReplyThread).Methods("POST")
    n := negroni.Classic()
    n.Use(negroni.HandlerFunc(BasicAuthMiddleware))
    n.Use(MongoMiddleware())
    // router goes last
    n.UseHandler(router)
    n.Run(":" + config["PORT"])
}

func ProcessInbound(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
    todoId := vars["todoId"]
    fmt.Fprintln(w, "Todo show:", todoId)
    fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}

func CreateThread(w http.ResponseWriter, r *http.Request) {
    var nMsg Message
    UnmarshalObject(r.Body, nMsg)
    thedb := context.Get(r, db).(*mgo.Database)
    tColl := thedb.C("message_threads")
    var owner = Owner{bson.ObjectIdHex(context.Get(r, userId).(string))}
    nThread := Thread{bson.NewObjectId(), owner, []Message{nMsg}}
    err := tColl.Insert(nThread)
    
    if err != nil {
        log.Fatal(err)
    }
    JSONResponse(w, nThread)
}

//gets all the threads for which the authenticated user is the owner
func GetAllThreads(w http.ResponseWriter, r *http.Request) {
    thedb := context.Get(r, db).(*mgo.Database)
    threads := []Thread{}
    tColl := thedb.C("message_threads")
    iter := tColl.Find(bson.M{"owner.id": bson.ObjectIdHex(context.Get(r, userId).(string))}).Limit(50).Iter()
    err := iter.All(&threads)
    if err != nil {
        log.Fatal(err)
    }
    JSONResponse(w, threads)
}

//return the requested thread, verifying the owner is the authenticated user
func GetOneThread(w http.ResponseWriter, r *http.Request) {
    tId := mux.Vars(r)["threadId"]
    //verify threadid is a valid objectid
    if !bson.IsObjectIdHex(tId) {
        http.Error(w, "NO auth header", http.StatusBadRequest)
        return
    }
    thedb := context.Get(r, db).(*mgo.Database)
    var thread Thread
    tColl := thedb.C("message_threads")
    err := tColl.Find(bson.M{"_id": bson.ObjectIdHex(tId),"owner.id": bson.ObjectIdHex(context.Get(r, userId).(string))}).One(&thread)
    if err != nil {
        http.Error(w, "NOT FOUND", http.StatusNotFound)
        return
    }

    JSONResponse(w, thread)
}

func ReplyThread(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}


func JSONResponse(w http.ResponseWriter, m interface{}) {
    j, err := json.Marshal(m)
    if err != nil {
        panic(err)
    }
    w.Header().Set("Content-Type", "application/json")
    w.Write(j)
}

func MongoMiddleware() negroni.HandlerFunc {
    database := "wure" //os.Getenv("DB_NAME")
    session, err := mgo.Dial(config["DBHOST"])
    
    if err != nil {
        panic(err)
    }
    session.SetMode(mgo.Monotonic, true)
    
    return negroni.HandlerFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
        reqSession := session.Clone()
        defer reqSession.Close()
        thedb := reqSession.DB(database)
        context.Set(r, db, thedb)
        next(rw, r)
    })
}

//exploits basic auth to auth the user making the request (MessageThread owner)
func BasicAuthMiddleware(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
    if len(r.Header["Authorization"]) == 0 {
        http.Error(rw, "NO auth header", http.StatusBadRequest)
        return
    }
    auth := strings.SplitN(r.Header["Authorization"][0], " ", 2)
    if len(auth) != 2 || auth[0] != "Basic" {
        http.Error(rw, "NO auth header", http.StatusBadRequest)
        return
    }
 
    payload, _ := base64.StdEncoding.DecodeString(auth[1])
    pair := strings.SplitN(string(payload), ":", 2)
    if len(pair) != 2 || !IsUserIdValid(pair[0]){
        http.Error(rw, "authorization failed", http.StatusUnauthorized)
        return
    }
    
    context.Set(r, userId, pair[0])
    next(rw, r)
    // do some stuff after
}
 

//this middleware marshals json
func UnmarshalObject(body io.Reader, obj interface{}) {
    decoder := json.NewDecoder(body)
    err := decoder.Decode(&obj)
    if err != nil {
        panic(err)
    }
}

//verifies the provided auth header values are actually valid 
func IsUserIdValid(uid string) bool {
    return bson.IsObjectIdHex(uid)
}

//loads the app configuration
func LoadConfig(fname string, config interface{}) {
    file, _ := os.Open(fname)
    defer file.Close()
    decoder := json.NewDecoder(file)
    err := decoder.Decode(config)
    if err != nil {
        fmt.Println("error loadin go config file:", err)
        panic(err)
    }
}