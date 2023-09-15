package main

import "fmt"

func main() {

	c := getFlags()
	Debug(c, c)
	client, err := buildKubeClient(c)
	if err != nil {
		panic(err.Error())
	}

	cl := NewCluster(c, client)
	err = cl.NewQueue("test")
	if err != nil {
		panic(err.Error())
	}

	q, err := cl.GetQueue("test")
	if err != nil {
		panic(err.Error())
	}

	h, err := q.Push([]byte("I've got a lovely bunch of coconuts!"))
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(h)
	fmt.Println(q.Length)

}
