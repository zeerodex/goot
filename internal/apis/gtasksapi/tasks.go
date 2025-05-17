package gtasksapi

import (
	"fmt"
	"log"

	gtasks "google.golang.org/api/tasks/v1"

	"github.com/zeerodex/goot/internal/tasks"
)

func GetTasksLists(srv *gtasks.Service) {
	r, err := srv.Tasklists.List().MaxResults(10).Do()
	if err != nil {
		log.Fatalf("Unable to retrieve task lists. %v", err)
	}

	fmt.Println("Task Lists:")
	if len(r.Items) > 0 {
		for _, i := range r.Items {
			fmt.Printf("%s (%s)\n", i.Title, i.Id)
		}
	} else {
		fmt.Print("No task lists found.")
	}
}

func GetTasksInDefault(srv *gtasks.Service) {
	r, err := srv.Tasks.List("@default").Do()
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Println("Task list in @default list:")
	for _, g := range r.Items {
		fmt.Println(tasks.ParseFromGtasks(g))
	}
}
