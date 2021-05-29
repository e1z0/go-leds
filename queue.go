package main

import (
        "container/heap"
        "fmt"
        "time"
        "errors"
)

var (
        priorityQueue    PriorityQueue
)

type PriorityQueue []*QueueItem

type QueueItem struct {
        Type    int
        Priority int
        Text     string
        Index    int // The index of the item in the heap.
}


func MainQueueLoop() {
        //forever := make(chan bool)
        // QUEUE HANDLER HERE
        priorityQueue = make(PriorityQueue, 0)
        heap.Init(&priorityQueue)
        go QueueHandler()
}




func AddToQueue(item QueueItem) {
        heap.Push(&priorityQueue, &item)
        if DEBUG > 0 {
                fmt.Printf("Queue added: %v\n", item)
        }
}


// Returns the lenght of the queue
func (pq PriorityQueue) Len() int { return len(pq) }

// Returns counts by the queues priorities
func (pq PriorityQueue) LenLevel(level int) int {
        count := 0
        for _, item := range pq {
                if item.Priority == level {
                        count++
                }
        }
        return count
}

// pq.PriorityQueue - picks the lowest expiration number
func (pq PriorityQueue) Less(i, j int) bool {
        // We want Pop to give us the lowest based on expiration number as the priority
        // The lower the expiry, the higher the priority
        return pq[i].Priority < pq[j].Priority
}

func (pq *PriorityQueue) update(item *QueueItem, typ int, priority int) {
        item.Type = typ
        item.Priority = priority
        heap.Fix(pq, item.Index)
}

// Push - adds the item to queue
func (pq *PriorityQueue) Push(x interface{}) {
        n := len(*pq)
        item := x.(*QueueItem)
        item.Index = n
        *pq = append(*pq, item)
}

// pq.PriorityQueue.Swap - swaps the upside down the queue list
func (pq PriorityQueue) Swap(i, j int) {
        pq[i], pq[j] = pq[j], pq[i]
        pq[i].Index = i
        pq[j].Index = j
}



func QueueHandler() {
        for {
                if priorityQueue.Len() > 0 {
                                item := heap.Pop(&priorityQueue).(*QueueItem)
                                fmt.Printf("Running QUEUE Type-> %d, Priority-> %d DATA->(%v)\n", item.Type, item.Priority, item.Text)
                                DISPLAY_INUSE=true
                                WriteMessage(item.Text)
                                DISPLAY_INUSE=false
                                }
                if Global.Debug {
                fmt.Printf("QUEUE TOTALS: -> %d\n", priorityQueue.LenLevel(2))
                }
                time.Sleep(1 * time.Second) // queue tick
        }
}



// Pop - Picks up the higher item in queue
func (pq *PriorityQueue) Pop() interface{} {
        old := *pq
        n := len(old)
        item := old[n-1]
        item.Index = -1
        *pq = old[0 : n-1]
        return item
}

// PopLevel2 - Pop level2 queue
func PopLevel2(pq *PriorityQueue) (*QueueItem, error) {
        pqas := *pq
        var item *QueueItem
        for i := 0; i < len(pqas); i++ {
                if pqas[i].Priority == 2 {
                        item = pqas[i]
                        pqas[i], pqas[len(pqas)-1] = pqas[len(pqas)-1], pqas[i]
                        pqas = pqas[:len(pqas)-1]
                        break
                }
        }
        priorityQueue = pqas

        if item == nil {
                err := errors.New("Queue list is empty for PopLevel2")
                return nil, err
        }

        return item, nil

}
