package main

import (
  "fmt"
  "strings"
  "regexp"
  "net/http"
  "io/ioutil"
  "os/signal"
  "os"
  "flag"
  "time"
)



/////////////////////////////////////////////
// Store
/////////////////////////////////////////////


type Store struct {
  Nodes map[string]*Store
  isTerminal bool
  lock Semaphore
}

// add that string to the store.
func (self *Store) insert(name string){
  self.Lock()
  if(len(name) == 0) {
    self.isTerminal = true
  } else {
    nextChar := strings.SplitN(name,"",2)[0]
    nextStore := self.Nodes[nextChar]
    if nextStore == nil {
      nextStore = &Store{make(map[string]*Store), false,make(Semaphore,1)}
      self.Nodes[nextChar] = nextStore
    }
    if len(name) == 1 {
      nextStore.isTerminal = true
    } else {
      self.Unlock()
      nextStore.insert(name[1:])
      return
    }
  }
  self.Unlock()
}

// prints out every string the store
func (self *Store) Print() {
  self.PrintString("")
}

// accumulatory helpter for store printer
func (self *Store) PrintString(acc string) {
  self.Lock()
  if self.isTerminal {
    fmt.Printf("%s\n",acc)
  }
  for c,s := range self.Nodes {
    s.PrintString(acc + c)
  }
  self.Unlock()
}

// lock this node of the store. does not lock its children
func (self *Store) Lock(){
  self.lock.Lock()
}

// unlock this node of the store. does not unlock its children
func (self *Store) Unlock(){
  self.lock.Unlock()
}

// get the number of strings in the store
func (self *Store) Size() int {
  return self.size(0)
}

// get the number of strings in the store using an accumulator
func (self *Store) size(acc int) int {
  if self.isTerminal {
    acc += 1
  }

  self.Lock()
  for _, node := range self.Nodes {
    acc = node.size(acc)
  }

  self.Unlock()

  return acc
}

// does this contain the given string?
func (self *Store) contain(s string) bool {
  if s == "" {
    return self.isTerminal
  }

  self.Lock()
  subnode := self.Nodes[s[0:1]]
  self.Unlock()
  if subnode != nil {
    return subnode.contain(s[1:])
  }
  return false
}
/////////////////////////////////////////////
// Semaphores
/////////////////////////////////////////////
type Semaphore chan bool

// acquire n resources
func (s Semaphore) Lock() {
  s<-false
}

// release n resources
func (s Semaphore) Unlock() {
  <-s
}

/////////////////////////////////////////////
// Links
/////////////////////////////////////////////

type Link struct {
  Url string
  Depth int
}

// gets the html from the given link
func (self *Link) UrlGet() string {
  var httpClient *http.Client = &http.Client{}
  for ;; {
    resp, err := httpClient.Get(self.Url)
    if err != nil {
      if IsDebugging {
        fmt.Printf("error getting page %s\n error was:%+v\n",self.Url,err)
      }
      if IsFailure(err) {
        return ""
      } else {
        time.Sleep(10 * time.Millisecond)
      }
    } else {
      body, _ := ioutil.ReadAll(resp.Body)
      resp.Body.Close()
      return string(body)
    }
  }
  // should never get here
  return ""
}

// is this an error not caused by request overload
func IsFailure(err error) bool {
  return !strings.Contains(err.Error(),"too many open files") && !strings.Contains(err.Error(),"Temporary failure in name resolution")
}

/////////////////////////////////////////////
// Main
/////////////////////////////////////////////
var ThreadCount int
var MaxSearchDepth int

var ThreadLocker Semaphore

var MainStore Store = Store{make(map[string]*Store),false,make(Semaphore,1)}

var StartLink *Link
var StartUrl string = "/wiki/Adolf_Hitler"
var NoRepeat bool = true;
var LinkRegex = "/wiki/.*"
var UrlPrefix = "http://en.wikipedia.org"
var IsDebugging bool = false

const ThreadCountDesc string = "specifies number of worker threads spawned"
const MaxSearchDesc   string = "specifies the search depth. < 0 will never terminate"
const StartUrlDesc string = "the path to start with"
const UrlPrefixDesc string = "the base url for all requests, including the first"
const NoRepeatDesc string = "Repeat links that have been seen before"
const LinkRegexDesc = "What regex should be used to match all links?"
const IsDebuggingDesc = "Enter Debug Mode"

var RXC = regexp.MustCompile

var TitleRegexp *regexp.Regexp = RXC("<title>.*<title>")
var MainDivHead string = "<div id=\"mw-content-text"
var MainDivEnd  string = "\n<!-- /bodyContent -->"
var LinkRegexp  *regexp.Regexp

func main() {

  ParseCommandLine()

  fmt.Printf("scanning %s\n",StartLink.Url)
  if NoRepeat {
    fmt.Printf("will not repeat links\n")
  } else {
    fmt.Printf("will repeat links\n")
  }

  if IsDebugging {
    StartInteruptHandler()
  }

  values := make([]*Link,1)
  values[0] = StartLink

  resultChannel := make(chan []*Link)

  for i := 0; i < MaxSearchDepth; i += 1 {
    start := time.Now()
    fmt.Printf("starting sweap of depth %d\n",i)
    go StartThreads(values,resultChannel)
    results := len(values)
    values = make([]*Link,0)
    for i := 0; i < results; i += 1 {
      newValues := <-resultChannel
      for _,l := range newValues {
        values = append(values,l)
      }
    }
    fmt.Printf("finish sweap of depth %d in %v, %d links found\n",i,time.Now().Sub(start),len(values))
  }

  //MainStore.Print()
  return
}

// gets used command line arguments
func ParseCommandLine() {
  flag.IntVar(&ThreadCount,"t",100,ThreadCountDesc)
  flag.StringVar(&UrlPrefix,"u",UrlPrefix,UrlPrefixDesc)
  flag.StringVar(&LinkRegex,"l",LinkRegex,LinkRegexDesc)
  flag.BoolVar(&IsDebugging,"debug",IsDebugging,IsDebuggingDesc)
  MaxSearchDepthFlag := flag.Int("d",3,MaxSearchDesc)
  StartUrlFlag       := flag.String("p",StartUrl,StartUrlDesc)
  NoRepeatFlag       := flag.Bool("r",false,NoRepeatDesc)

  flag.Parse()

  MaxSearchDepth = *MaxSearchDepthFlag + 1
  StartUrl = UrlPrefix + *StartUrlFlag
  StartLink = &Link{StartUrl,0}
  NoRepeat = !*NoRepeatFlag

  LinkRegexp = RXC("<a href=\"" + LinkRegex + "\".*>.*</a>")

  ThreadLocker = make(Semaphore,ThreadCount)
}

// start interupt handler
func StartInteruptHandler(){
  go func () {
    var interuptc chan os.Signal = make(chan os.Signal,1)
      signal.Notify(interuptc, os.Interrupt)
      <-interuptc
      panic(fmt.Sprintf("Showing stack traces\n"))
  }()
}

// Start threads for each link
func StartThreads(values []*Link, ret chan []*Link){
  for _,l := range values {
    StartCrawler(l,ret)
  }
}

// Attempts to start a new crawler thread. 
// Blocks if maximum thread count has been reached
func StartCrawler(NextLink *Link, ret chan []*Link){
  ThreadLocker.Lock()
  go StartThread(NextLink,ret)
}

// basic worker thread.
func StartThread(NextLink *Link, ret chan []*Link){
    body := NextLink.UrlGet()
    title := TitleGet(body)
    HandleNewLink(NextLink,title,body,ret)
}

// function for parsing a link
func HandleNewLink(L *Link, title string, body string, ret chan []*Link) {
  if L.Depth != MaxSearchDepth {
    Links := getLinks(body,L.Depth)
    if IsDebugging && len(Links) == 0{
      fmt.Printf("No links found for %s\n",L.Url)
      fmt.Printf("Body was:\n%s",body)
    }
    if NoRepeat {
      Links = PruneDups(Links)
    }
    for _,l := range Links {
      MainStore.insert(l.Url)
    }
    ret<-Links
    ThreadLocker.Unlock()
  }
}

// remove all elements already seen in the main store
func PruneDups(links []*Link) []*Link {
  ret := make([]*Link,0)
  for _,l := range links {
    if !MainStore.contain(l.Url) {
      ret = append(ret,l)
    }
  }
  return ret
}

/////////////////////////////////////////////
// utilities
/////////////////////////////////////////////
// get the title from the wikipedia page
func TitleGet(body string) string {
  return TitleRegexp.FindString(body)
}

// get all Links from the content of the body of a page
func getLinks(body string, currentdepth int) []*Link {
  content := getContent(body)
  depth := currentdepth + 1
  links   := LinkRegexp.FindAllString(content,-1)
  var retLinks []*Link = make([]*Link,len(links))
  for i,s := range links {
    name := UrlPrefix + strings.SplitN(s,"\"",3)[1]
    retLinks[i] = &Link{name,depth}
  }
  return retLinks
}

// get the content div from a wikipedia page
func getContent(body string) string {
  //start := strings.Index(body,MainDivHead)
  //end   := strings.Index(body,MainDivEnd)
  //fmt.Printf("splitting body at %d,%d",start,end)
  //return body[start:end]
  return body
}
