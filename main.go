package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv" // Ajout de l'importation manquante
)

// Structures de données
type ContactForm struct {
	Name    string
	Email   string
	Message string
	Errors  map[string]string
}

type TemplateData struct {
	Form    ContactForm
	Success bool
}

func main() {
	// Chargement des variables d'environnement
	err := godotenv.Load()
	if err != nil {
		log.Println("Aucun fichier .env trouvé - utilisation des variables système")
	}

	// Configurer les routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/contact", contactHandler)

	// Servir les fichiers statiques
	staticDir := filepath.Join("front-end", "static")
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	// Démarrer le serveur
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}
	log.Printf("Serveur démarré sur http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

// Handlers HTTP
func homeHandler(w http.ResponseWriter, r *http.Request) {
	tmplPath := filepath.Join("front-end", "templates", "index.html")

	success := r.URL.Query().Get("success") == "true"
	data := TemplateData{
		Form:    ContactForm{},
		Success: success,
	}

	renderTemplate(w, tmplPath, data)
}

func contactHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Parser le formulaire
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Erreur lors de la lecture du formulaire", http.StatusBadRequest)
		return
	}

	form := ContactForm{
		Name:    strings.TrimSpace(r.FormValue("name")),
		Email:   strings.TrimSpace(r.FormValue("email")),
		Message: strings.TrimSpace(r.FormValue("message")),
		Errors:  make(map[string]string),
	}

	// Validation
	if form.Name == "" {
		form.Errors["Name"] = "Le nom est requis"
	}
	if form.Email == "" {
		form.Errors["Email"] = "L'email est requis"
	} else if !strings.Contains(form.Email, "@") {
		form.Errors["Email"] = "Email invalide"
	}
	if form.Message == "" {
		form.Errors["Message"] = "Le message est requis"
	}

	// Si erreurs, réafficher le formulaire
	if len(form.Errors) > 0 {
		tmplPath := filepath.Join("front-end", "templates", "index.html")
		renderTemplate(w, tmplPath, TemplateData{Form: form})
		return
	}

	// Rediriger vers le mailto: avec les informations pré-remplies
	mailto := fmt.Sprintf("mailto:siondigitale@gmail.com?subject=Message%%20de%%20%s&body=%s", 
		strings.ReplaceAll(form.Name, " ", "%20"),
		strings.ReplaceAll(form.Message, " ", "%20"))

	http.Redirect(w, r, mailto, http.StatusSeeOther)
}

// Fonctions utilitaires
func renderTemplate(w http.ResponseWriter, tmplPath string, data interface{}) {
	// Vérifier que le fichier existe
	if _, err := os.Stat(tmplPath); os.IsNotExist(err) {
		http.Error(w, "Template introuvable", http.StatusInternalServerError)
		return
	}

	// Parser le template avec gestion des erreurs
	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erreur de parsing du template: %v", err), http.StatusInternalServerError)
		return
	}

	// Exécuter le template
	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Erreur d'exécution du template: %v", err), http.StatusInternalServerError)
	}
}