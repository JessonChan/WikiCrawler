package main
import (
  "strings"
  "fmt"
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


