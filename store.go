package main
import (
  "strings"
  "fmt"
)
/////////////////////////////////////////////
// Store
/////////////////////////////////////////////


type Store struct {
  nodes map[string]*Store
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
    nextStore := self.nodes[nextChar]
    if nextStore == nil {
      nextStore = &Store{make(map[string]*Store), false,make(Semaphore,1)}
      self.nodes[nextChar] = nextStore
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
  strs := self.Iterate()
  for url := <-strs; url != ""; url = <-strs {
    fmt.Printf("%s\n",url)
  }
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

func (self *Store) Join(other *Store) {
  strs := other.Iterate()
  for url :=<-strs; url != ""; url =<-strs {
    self.Insert(url)
  }
}

// get the number of strings in the store using an accumulator
func (self *Store) size(acc int) int {
  if self.isTerminal {
    acc += 1
  }

  self.Lock()
  for _, node := range self.nodes {
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
  subnode := self.nodes[s[0:1]]
  self.Unlock()
  if subnode != nil {
    return subnode.contain(s[1:])
  }
  return false
}

// iterates through this store putting each string into the channel
// puts one emtpy sting in the chan before terminating
func (self *Store) Iterate() chan string {
  output := make(chan string)
  go self.startIterate("",output)
  return output;
}

func (self *Store) startIterate(acc string, output chan string) {
  self.iterate(acc,output)
  output<-""
}

func (self *Store) iterate(acc string, output chan string) {
  if self.isTerminal {
    output<-acc
  }

  self.Lock()
  for c,s := range self.nodes {
    next := acc + c
    self.Unlock()
    s.iterate(next,output)
    self.Lock()
  }
  self.Unlock()
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
