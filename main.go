package main

import (
	"fmt"
	"sync"
	"time"
)

type Consumer struct {
	msgs *chan int
}

// NewConsumer creates a Consumer
func NewConsumer(msgs *chan int) *Consumer {
	return &Consumer{msgs: msgs}
}

// consume reads the msgs channel
func (c *Consumer) consume() {
	fmt.Println("consume: Started")
	println("Varun")
	for {
		msg := <-*c.msgs
		fmt.Println("consume: Received:", msg)
	}
}

// Producer definition
type Producer struct {
	msgs *chan int
	done *chan bool
}

// NewProducer creates a Producer
func NewProducer(msgs *chan int, done *chan bool) *Producer {
	return &Producer{msgs: msgs, done: done}
}

// produce creates and sends the message through msgs channel
func (p *Producer) produce(max int) {
	fmt.Println("produce: Started")
	for i := 0; i < max; i++ {
		fmt.Println("produce: Sending ", i)
		*p.msgs <- i
	}
	*p.done <- true // signal when done
	fmt.Println("produce: Done")
}

const (
	sleeping = iota
	checking
	cutting
)

var stateLog = map[int]string{
	0: "Sleeping",
	1: "Checking",
	2: "Cutting",
}
var wg *sync.WaitGroup // Amount of potentional customers

type Barber struct {
	name string
	sync.Mutex
	state    int // Sleeping/Checking/Cutting
	customer *Customer
}

type Customer struct {
	name string
}

func (c *Customer) String() string {
	return fmt.Sprintf("%p", c)[7:]
}

func NewBarber() (b *Barber) {
	return &Barber{
		name:  "Sam",
		state: sleeping,
	}
}

// Barber goroutine
// Checks for customers
// Sleeps - wait for wakers to wake him up
func barber(b *Barber, wr chan *Customer, wakers chan *Customer) {
	for {
		b.Lock()
		defer b.Unlock()
		b.state = checking
		b.customer = nil

		// checking the waiting room
		fmt.Printf("Checking waiting room: %d\n", len(wr))
		time.Sleep(time.Millisecond * 100)
		select {
		case c := <-wr:
			HairCut(c, b)
			b.Unlock()
		default: // Waiting room is empty
			fmt.Printf("Sleeping Barber - %s\n", b.customer)
			b.state = sleeping
			b.customer = nil
			b.Unlock()
			c := <-wakers
			b.Lock()
			fmt.Printf("Woken by %s\n", c)
			HairCut(c, b)
			b.Unlock()
		}
	}
}

func DeferCheck() {

	fmt.Println("counting")

	for i := 0; i < 10; i++ {
		defer fmt.Println(i)
	}

	fmt.Println("done")
}

func split(sum int) (x, y int) {
	x = sum * 4 / 9
	y = sum - x
	return
}

func HairCut(c *Customer, b *Barber) {
	b.state = cutting
	b.customer = c
	b.Unlock()
	fmt.Printf("Cutting  %s hair\n", c)
	time.Sleep(time.Millisecond * 100)
	b.Lock()
	wg.Done()
	b.customer = nil
}

// customer goroutine
// just fizzles out if it's full, otherwise the customer
// is passed along to the channel handling it's haircut etc
func customer(c *Customer, b *Barber, wr chan<- *Customer, wakers chan<- *Customer) {
	// arrive
	time.Sleep(time.Millisecond * 50)
	// Check on barber
	b.Lock()
	fmt.Printf("Customer %s checks %s barber | room: %d, w %d - customer: %s\n",
		c, stateLog[b.state], len(wr), len(wakers), b.customer)
	switch b.state {
	case sleeping:
		select {
		case wakers <- c:
		default:
			select {
			case wr <- c:
			default:
				wg.Done()
			}
		}
	case cutting:
		select {
		case wr <- c:
		default: // Full waiting room, leave shop
			wg.Done()
		}
	case checking:
		panic("Customer shouldn't check for the Barber when Barber is Checking the waiting room")
	}
	b.Unlock()
}

func main() {
	/*
	   // profile flags
	   cpuprofile := flag.String("cpuprofile", "", "write cpu profile to `file`")
	   memprofile := flag.String("memprofile", "", "write memory profile to `file`")

	   // get the maximum number of messages from flags
	   max := flag.Int("n", 5, "defines the number of messages")

	   flag.Parse()

	   // utilize the max num of cores available
	   runtime.GOMAXPROCS(runtime.NumCPU())

	   // CPU Profile
	   if *cpuprofile != "" {
	   f, err := os.Create(*cpuprofile)
	   if err != nil {
	   log.Fatal("could not create CPU profile: ", err)
	   }
	   if err := pprof.StartCPUProfile(f); err != nil {
	   log.Fatal("could not start CPU profile: ", err)
	   }
	   defer pprof.StopCPUProfile()
	   }

	   var msgs = make(chan int)  // channel to send messages
	   var done = make(chan bool) // channel to control when production is done

	   // Start a goroutine for Produce.produce
	   go NewProducer(&msgs, &done).produce(*max)

	   // Start a goroutine for Consumer.consume
	   go NewConsumer(&msgs).consume()

	   // Finish the program when the production is done
	   <-done

	   // Memory Profile
	   if *memprofile != "" {
	   f, err := os.Create(*memprofile)
	   if err != nil {
	   log.Fatal("could not create memory profile: ", err)
	   }
	   runtime.GC() // get up-to-date statistics
	   if err := pprof.WriteHeapProfile(f); err != nil {
	   log.Fatal("could not write memory profile: ", err)
	   }
	   f.Close()
	   }


	*/

	b := NewBarber()
	b.name = "Rocky"
	WaitingRoom := make(chan *Customer, 5) // 5 chairs
	Wakers := make(chan *Customer, 1)      // Only one waker at a time
	go barber(b, WaitingRoom, Wakers)

	time.Sleep(time.Millisecond * 100)
	wg = new(sync.WaitGroup)
	n := 10
	wg.Add(10)
	// Spawn customers
	for i := 0; i < n; i++ {
		time.Sleep(time.Millisecond * 50)
		c := new(Customer)
		go customer(c, b, WaitingRoom, Wakers)
	}

	wg.Wait()
	fmt.Println("No more customers for the day")

	//DeferCheck()
	//println(split(15))

	func(l int, b int) {
		fmt.Println(l * b)
	}(20, 30)

}
