package main

import (
	"context"
	"embed-static-files/internal/box"
	"embed-static-files/internal/config"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"text/template"
	"time"
)

type PageData struct {
	Title       string
	Heading     string
	Description string
}

func main() {

	cfgPath, err := config.ParseFlags()
	if err != nil {
		log.Fatal(err)
	}

	// Set up a channel to listen to for interrupt signals
	var runChan = make(chan os.Signal, 1)

	config, err := config.NewConfig(cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	// Set up a context to allow for graceful server shutdowns in the event
	// of an OS interrupt (defers the cancel just in case)
	ctx, cancel := context.WithTimeout(
		context.Background(),
		config.Server.Timeout.Server,
	)
	defer cancel()

	// Define embed file (for a short)
	index := string(box.Get("/index.html"))

	// Template
	tmpl := template.Must(template.New("").Parse(index))

	// Handle function
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := PageData{
			Title:       "The easiest way to embed static files into a binary file",
			Heading:     "This is easiest way",
			Description: "My life credo is 'If you can not use N, do not use N'.",
		}

		// Execute template with data
		if err := tmpl.Execute(w, data); err != nil {
			log.Fatal(err)
		}
	})

	// Define server options
	server := &http.Server{
		Addr:         config.Server.Host + ":" + config.Server.Port,
		Handler:      nil,
		ReadTimeout:  config.Server.Timeout.Read * time.Second,
		WriteTimeout: config.Server.Timeout.Write * time.Second,
		IdleTimeout:  config.Server.Timeout.Idle * time.Second,
	}

	// Handle ctrl+c/ctrl+x interrupt
	signal.Notify(runChan, os.Interrupt, syscall.SIGTSTP)
	// Alert the user that the server is starting
	log.Printf("Server is starting on %s\n", server.Addr)

	// Run the server on a new goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				// Normal interrupt operation, ignore
			} else {
				log.Fatalf("Server failed to start due to err: %v", err)
			}
		}
	}()

	// Block on this channel listeninf for those previously defined syscalls assign
	// to variable so we can let the user know why the server is shutting down
	interrupt := <-runChan

	// // Start server
	// if err := http.ListenAndServe(":8080", nil); err != nil {
	// 	log.Fatal(err)
	// }

	// If we get one of the pre-prescribed syscalls, gracefully terminate the server
	// while alerting the user
	log.Printf("Server is shutting down due to %+v\n", interrupt)
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server was unable to gracefully shutdown due to err: %+v", err)
	}
}
