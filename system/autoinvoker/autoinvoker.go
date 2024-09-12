package autoinvoker

import (
	"bytes"
	"encoding/json"
	"event_horizon/system/db"
	"event_horizon/templates"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/exec"

	"math/rand"
)

var RedirectionTarget string

func AutoInitialize(as_domain string, context *map[string]string, requires []string, mode string, redirect_target string) {
	if len(os.Args) == 1 {
		RedirectionTarget = redirect_target
		if mode == "dev" {
			auto_server("8001")
		} else {
			var ip string
			if mode == "local" {
				ip = "127.0.0.1"
			} else {
				ip = get_remote_ip(os.Getenv("VERSE_REFLECTOR"))
			}
			port := get_random_port()
			if self_declare(as_domain, ip, port, "domain_workers") {
				fmt.Printf("Self declared as %s with %s:%s in domain_workers\n", as_domain, ip, port)
				auto_server(port)
			} else {
				fmt.Println("Error initializing context.")
				os.Exit(1)
			}
		}

	} else if len(os.Args) == 2 {
		arg := os.Args[1]
		err := json.Unmarshal([]byte(arg), &context)
		if err != nil {
			fmt.Println("Error parsing JSON:", err)
			return
		}
		for _, key := range requires {
			if _, exists := (*context)[key]; !exists {
				fmt.Print("BAD_REQUEST")
				os.Exit(1)
			}
		}
	} else {
		fmt.Printf("Bad Arguments.")
	}

}

func self_declare(as_domain string, ip string, port string, collection string) bool {
	context := make(map[string]string)
	context["domain"] = as_domain
	context["ip"] = ip
	context["port"] = port

	if db.Connect() {
		if db.CreateCollection(collection) {
			if db.UpsertItemInCollection(collection, context, "domain") {
				return true
			}
		}
	}
	return false
}
func auto_server(port string) {
	fs := http.FileServer(http.Dir("templates/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	fmt.Printf("Server listening on port %s on all interfaces.\n", port)
	http.HandleFunc("/", auto_handler)
	err := http.ListenAndServe("0.0.0.0:"+port, nil)
	if err != nil {
		fmt.Println("Error starting server:", err)
	}
}
func auto_handler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		auto_get(w)
	case http.MethodPost:
		handle_impulse(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}
func auto_get(w http.ResponseWriter) {
	tmpl, _ := template.ParseFiles("templates/index.html")
	err := tmpl.Execute(w, nil)
	if err != nil {
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
}
func handle_impulse(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	context := make(map[string]string)
	for key, values := range r.Form {
		context[key] = values[0]
	}

	fmt.Println("Context Received: ")
	for k, v := range context {
		fmt.Printf("%s : %s\n", k, v)
	}

	contextJSON, err := json.Marshal(context)
	if err != nil {
		http.Error(w, "Error encoding JSON", http.StatusInternalServerError)
		return
	}

	outbuf, response := auto_invoke(contextJSON)
	if response != nil {
		tmpl := template.New("response").Funcs(templates.GetTemplateFunctions())
		tmpl, _ = tmpl.ParseFiles("templates/response.html")
		err := tmpl.Execute(w, outbuf)
		if err != nil {
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
	} else {
		http.Redirect(w, r, RedirectionTarget, http.StatusSeeOther)
	}

}

func auto_invoke(jsonContext []byte) (string, error) {
	cmd := exec.Command(os.Args[0], string(jsonContext))

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	return outBuf.String(), err
}

func get_remote_ip(url string) string {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error making request to ip mirror:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading ip mirror response:", err)
		os.Exit(1)
	}
	return string(body)
}

func get_random_port() string {
	port := rand.Intn(65535-1024) + 1024
	return fmt.Sprintf("%d", port)
}
