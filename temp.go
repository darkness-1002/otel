package main

type response struct {
	output interface{}
	error  error
}

type demographics struct {
	age int
	//other demographics data
}

func cache(refId string, ch chan response) {
	// do some work
	// send result on channel

	return
}

type SolrDocs struct {
	refId string
	uid   string
	edi   string
	score float32
}

// txId string , demographics *Demographics,queryType string,maxCandidates int
func getSolr(d demographics, ch chan response) {
	// do some work
	// send result on channel

	return
}
func main() {
	//get packet
	//extract refId
	refId := "123"
	ch := make(chan response)
	go cache(refId, ch)
	x := <-ch
	if x.error != nil {
		// handle error
	}
	input, ok := x.output.(demographics)
	if !ok {
		// handle error
	}

	go getSolr(input, ch)
	// go func() {
	// 	// do some work
	// 	// send result on channel
	// 	getSolr()
	// }()
	y := <-ch
	if y.error != nil {
		// handle error
	}
	input2, ok := y.output.(SolrDocs)
	if !ok {
		// handle error
	}
	println(input2.refId)

}
