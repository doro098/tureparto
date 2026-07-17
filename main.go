package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// verifyToken devuelve el token de verificación desde la variable de entorno.
// Debes configurar este mismo token en Meta WhatsApp Cloud API.
func verifyToken() string {
	token := os.Getenv("VERIFY_TOKEN")
	if token == "" {
		// Token por defecto para desarrollo. CÁMBIALO en producción.
		return "tureparto_token_seguro_123"
	}
	return token
}

// Estructura para parsear los mensajes entrantes de WhatsApp
type whatsappMessage struct {
	Entry []struct {
		Changes []struct {
			Value struct {
				Messages []struct {
					From string `json:"from"`
					Text struct {
						Body string `json:"body"`
					} `json:"text"`
				} `json:"messages"`
			} `json:"value"`
		} `json:"changes"`
	} `json:"entry"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	// Inicializar base de datos SQLite
	if err := initDB(); err != nil {
		log.Fatalf("❌ Error iniciando base de datos: %v", err)
	}
	defer db.Close()

	http.HandleFunc("/webhook", webhookHandler)
	http.HandleFunc("/", indexHandler)

	log.Printf("🚀 Servidor iniciado en http://localhost:%s", port)
	log.Printf("🔗 GET  /webhook → Verificación con Meta")
	log.Printf("📩 POST /webhook → Recibir mensajes de WhatsApp")
	log.Printf("🔑 Verify Token: %s (configúralo en Meta)", verifyToken())
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	fmt.Fprintf(w, "✅ Servidor de WhatsApp Webhook funcionando!\n")
	fmt.Fprintf(w, "📌 Endpoint: /webhook\n")
}

// webhookHandler maneja tanto la verificación (GET) como los mensajes entrantes (POST).
func webhookHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleVerification(w, r)
	case http.MethodPost:
		handleMessage(w, r)
	default:
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
	}
}

// handleVerification procesa la solicitud de verificación de Meta.
// Meta envía: GET /webhook?hub.mode=subscribe&hub.verify_token=<token>&hub.challenge=<challenge>
func handleVerification(w http.ResponseWriter, r *http.Request) {
	mode := r.URL.Query().Get("hub.mode")
	token := r.URL.Query().Get("hub.verify_token")
	challenge := r.URL.Query().Get("hub.challenge")
	tokenEsperado := verifyToken()

	log.Printf("📥 Verificación recibida: mode=%s, token=%s", mode, token)

	if mode == "subscribe" && token == tokenEsperado {
		log.Println("✅ Verificación exitosa! Meta ha confirmado el webhook.")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, challenge)
		return
	}

	log.Printf("❌ Verificación fallida: token incorrecto (esperado: %s)", tokenEsperado)
	http.Error(w, "Token de verificación inválido", http.StatusForbidden)
}

// handleMessage procesa los mensajes entrantes de WhatsApp y los guarda en BD.
func handleMessage(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("❌ Error leyendo body: %v", err)
		http.Error(w, "Error leyendo request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Parsear el mensaje de WhatsApp
	var msg whatsappMessage
	if err := json.Unmarshal(body, &msg); err != nil {
		log.Printf("📩 Mensaje recibido (no-WhatsApp o raw): %s", string(body))
	} else {
		// Mostrar el JSON bonito para depuración
		prettyJSON, _ := json.MarshalIndent(msg, "", "  ")
		log.Printf("📩 Webhook recibido:\n%s", string(prettyJSON))

		// Extraer, mostrar y guardar cada mensaje
		for _, entry := range msg.Entry {
			for _, change := range entry.Changes {
				for _, message := range change.Value.Messages {
					log.Printf("💬 De: %s | Mensaje: %s", message.From, message.Text.Body)

					// Guardar en base de datos
					if err := saveMessage(message.From, message.Text.Body); err != nil {
						log.Printf("❌ Error al guardar mensaje en BD: %v", err)
					}
				}
			}
		}
	}

	// Meta espera un 200 OK para confirmar que recibimos el mensaje
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}
