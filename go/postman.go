package main

import (
    "os"
    "fmt"
    "html"
    "log"
    "net/http"
    "io"
    "flag"
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
    //load configuration
    LoadConfig(&config)
    router := mux.NewRouter().StrictSlash(true)
    n := negroni.Classic()
    //setup routing and middleware     
    PrepareRouting(router, n)
    //and run
    n.Run(":" + config["PORT"])
}

//associates the routes to the router
func PrepareRouting(rt *mux.Router, n *negroni.Negroni){
    //routes
    rt.HandleFunc("/inbound", ProcessInbound).Methods("POST")
    rt.HandleFunc("/threads", CreateThread).Methods("POST")     //OK
    rt.HandleFunc("/threads", GetAllThreads).Methods("GET")     //OK
    rt.HandleFunc("/threads/{threadId}", GetOneThread).Methods("GET")   //OK
    rt.HandleFunc("/threads/{threadId}/reply", ReplyThread).Methods("POST") //OK
    //some middleware
    n.Use(negroni.HandlerFunc(BasicAuthMiddleware))
    n.Use(MongoMiddleware())
    // router goes last
    n.UseHandler(rt)
}

func ProcessInbound(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
    todoId := vars["todoId"]
    fmt.Fprintln(w, "Todo show:", todoId)
    fmt.Fprintf(w, "Hello, %q", html.EscapeString(r.URL.Path))
}

func CreateThread(w http.ResponseWriter, r *http.Request) {
    var nMsg Message
    err := UnmarshalObject(r.Body, &nMsg)
    if err != nil{
        http.Error(w, "Your JSON is not GOOD", http.StatusBadRequest)
        return
    }
    thedb := context.Get(r, db).(*mgo.Database)
    tColl := thedb.C("message_threads")
    var owner = Owner{bson.ObjectIdHex(context.Get(r, userId).(string))}
    nThread := Thread{bson.NewObjectId(), owner, []Message{nMsg}}
    err = tColl.Insert(nThread)
    if err != nil {
        log.Fatal(err)
    }
    //actually send out the mail
    //go NewMailProvider(config).SendMail(nThread.Id.String(), nMsg.From, []string{nMsg.To}, nMsg.Msg)
    //config thread creation
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
    tId := mux.Vars(r)["threadId"]
    //verify threadid is a valid objectid
    if !bson.IsObjectIdHex(tId) {
        http.Error(w, "NO auth header", http.StatusBadRequest)
        return
    }
    //unmarshal the new message to add
    var nMsg Message
    err := UnmarshalObject(r.Body, &nMsg)
    if err != nil{
        http.Error(w, "Your JSON is not GOOD", http.StatusBadRequest)
        return
    }
    //now let's find out to who we need to send the reply to, avoiding loops
    var thread Thread
    tColl := context.Get(r, db).(*mgo.Database).C("message_threads")
    err = tColl.Find(bson.M{"_id": bson.ObjectIdHex(tId),"owner.id": bson.ObjectIdHex(context.Get(r, userId).(string))}).One(&thread)
    if err != nil {
        http.Error(w, "Can't get your thread", http.StatusInternalServerError)
        return
    }
    for i := len(thread.Messages)-1; i >= 0; i-- {
        if thread.Messages[i].From != nMsg.From {
            nMsg.To = thread.Messages[i].From
            break
        }
    }
    //found?
    if nMsg.To == "" {
        http.Error(w, "Can't do this, loop will be", http.StatusInternalServerError)
        return
    }
    //ready for update
    err = tColl.Update(bson.M{"_id": bson.ObjectIdHex(tId), "owner.id": bson.ObjectIdHex(context.Get(r, userId).(string))}, bson.M{"$push": bson.M{"messages": nMsg}})
    if err != nil {
        if err == mgo.ErrNotFound{
            http.Error(w, "NOT FOUND", http.StatusNotFound)
            return
        }
        log.Fatal("Can't update document %v\n", err)
        http.Error(w, "Can't update document", http.StatusInternalServerError)
        return
    }
    //everything ok, send the mail
    //go NewMailProvider(config).SendMail(thread.Id.String(), nMsg.From, []string{nMsg.To}, nMsg.Msg)

    // return the updated thread
    thread.Messages[len(thread.Messages)] = nMsg
    JSONResponse(w, thread)
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
 

//this unmarshals json
func UnmarshalObject(body io.Reader, obj interface{}) error{
    return json.NewDecoder(body).Decode(obj)
}

//verifies the provided auth header values are actually valid 
func IsUserIdValid(uid string) bool {
    return bson.IsObjectIdHex(uid)
}

//loads the app configuration
func LoadConfig(config interface{}) {
    fname := flag.String("c", "conf_debug.json", "path to JSON config file")
    flag.Parse()
    file, _ := os.Open(*fname)
    defer file.Close()
    decoder := json.NewDecoder(file)
    err := decoder.Decode(config)
    if err != nil {
        fmt.Println("error loading go config file:", err)
        panic(err)
    }
}

