package main

import (
	"crypto/sha1"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

const (
	DefaultRefreshRate = 900 // Default image duration in seconds
	DefaultSecretKey   = "TRMNL"
	FriendlyIDLength   = 6 // Length of friendly ID
)

// Build information (set via ldflags)
var (
	Version   = "dev"
	BuildTime = "unknown"
	Commit    = "unknown"
)

type Config struct {
	Bind     string
	Port     int
	ImageDir string
	Setup    bool
}

type DisplayResponse struct {
	Status          int    `json:"status"`
	ImageURL        string `json:"image_url,omitempty"`
	Filename        string `json:"filename,omitempty"`
	RefreshRate     int    `json:"refresh_rate"`
	ResetFirmware   bool   `json:"reset_firmware"`
	UpdateFirmware  bool   `json:"update_firmware"`
	FirmwareURL     string `json:"firmware_url,omitempty"`
	SpecialFunction string `json:"special_function,omitempty"`
	Action          string `json:"action,omitempty"`
	Error           string `json:"error,omitempty"`
}

type SetupResponse struct {
	Status     int    `json:"status"`
	APIKey     string `json:"api_key"`
	FriendlyID string `json:"friendly_id"`
	ImageURL   string `json:"image_url"`
	Message    string `json:"message"`
}

type LogRequest struct {
	Log interface{} `json:"log"`
}

type DeviceState struct {
	CurrentImageIndex int
	LastUpdate        time.Time
}

type Server struct {
	config      Config
	deviceState map[string]*DeviceState
	stateMutex  sync.RWMutex
	images      []string
	imagesMutex sync.RWMutex
}

func main() {
	// Parse command line arguments first (for --help)
	config := parseArgs()

	// Write out a warning if SECRET_KEY_BASE is not set
	if os.Getenv("SECRET_KEY_BASE") == "" {
		log.Println("Warning: SECRET_KEY_BASE environment variable is not set. Unauthorized clients will be able to fetch screens.")
	}

	server := &Server{
		config:      config,
		deviceState: make(map[string]*DeviceState),
	}

	// Load images from directory
	if err := server.loadImages(); err != nil {
		log.Fatalf("Failed to load images: %v", err)
	}

	// Set up routes
	r := mux.NewRouter()
	r.Use(server.loggingMiddleware)
	r.Use(server.corsMiddleware)

	// API routes
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/display", server.handleDisplay).Methods("GET", "OPTIONS")
	api.HandleFunc("/setup", server.handleSetup).Methods("GET", "OPTIONS")
	api.HandleFunc("/setup/", server.handleSetup).Methods("GET", "OPTIONS")
	api.HandleFunc("/log", server.handleLog).Methods("POST", "OPTIONS")

	// Image serving route
	r.HandleFunc("/images/{filename}", server.handleImage).Methods("GET")

	addr := fmt.Sprintf("%s:%d", config.Bind, config.Port)
	log.Printf("TRMNL API Server %s (built %s, commit %s)", Version, BuildTime, Commit)
	log.Printf("Server starting on %s", addr)
	log.Printf("Serving images from: %s", config.ImageDir)
	log.Printf("Found %d images", len(server.images))
	if config.Setup {
		log.Printf("Device provisioning enabled via --setup flag")
	} else {
		log.Printf("Device provisioning disabled - use --setup flag to enable")
	}

	log.Fatal(http.ListenAndServe(addr, r))
}

func parseArgs() Config {
	var config Config

	// Define flags
	portFlag := flag.Int("port", 3000, "Port to listen on")
	bindFlag := flag.String("bind", "0.0.0.0", "Address to bind to")
	setupFlag := flag.Bool("setup", false, "Allow device provisioning via /api/setup endpoint")
	helpFlag := flag.Bool("help", false, "Show help message")

	flag.Parse()

	if *helpFlag {
		fmt.Printf("TRMNL API Server %s\n\n", Version)
		fmt.Println("Usage: trmnld [options] [image-directory]")
		fmt.Println("\nOptions:")
		flag.PrintDefaults()
		fmt.Println("\nEnvironment Variables:")
		fmt.Println("  SECRET_KEY_BASE    Optional, highly recommended. Used for device API key generation.")
		fmt.Println("\nArguments:")
		fmt.Println("  image-directory    Directory containing images (default: current directory)")
		os.Exit(0)
	}

	config.Port = *portFlag
	config.Bind = *bindFlag
	config.Setup = *setupFlag

	// Get image directory from args or use current directory
	args := flag.Args()
	if len(args) > 0 {
		config.ImageDir = args[0]
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			log.Fatalf("Could not get current working directory: %v", err)
		}
		config.ImageDir = cwd
	}

	return config
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Access-Token, ID, Battery-Voltage, FW-Version, RSSI, Height, Width, Special-Function")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s - %v", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start))
	})
}


func (s *Server) handleSetup(w http.ResponseWriter, r *http.Request) {
	macAddress := r.Header.Get("ID")
	if macAddress == "" {
		s.sendJSONResponse(w, SetupResponse{
			Status:  404,
			Message: "MAC address required in ID header",
		})
		return
	}

	// Check if setup is enabled
	if !s.config.Setup {
		log.Printf("Setup attempt from MAC address %s denied - setup flag not set", macAddress)
		s.sendJSONResponse(w, DisplayResponse{
			Status:        500,
			ResetFirmware: false,
			Error:         "Setup not enabled",
		})
		return
	}

	log.Printf("MAC address %s was authenticated", macAddress)

	apiKey := generateAPIKey(macAddress)

	// Generate friendly ID (first 6 chars of API key, uppercase, with dash)
	friendlyID := strings.ToUpper(apiKey[:3]) + "-" + strings.ToUpper(apiKey[3:FriendlyIDLength])

	s.sendJSONResponse(w, SetupResponse{
		Status:     200,
		APIKey:     apiKey,
		FriendlyID: friendlyID,
		Message:    fmt.Sprintf("Device registered with friendly ID '%s'", friendlyID),
		ImageURL:   fmt.Sprintf("%s://%s/images/%s", r.URL.Scheme, r.Host, s.images[0]),
	})
}

func generateAPIKey(mac string) string {
	secretKeyBase := os.Getenv("SECRET_KEY_BASE")
	if secretKeyBase == "" {
		secretKeyBase = DefaultSecretKey
	}

	hash := sha1.Sum([]byte(mac + secretKeyBase))
	return fmt.Sprintf("%x", hash)
}

func validateAPIKey(mac string, token string) bool {
	return generateAPIKey(mac) == token
}

func (s *Server) handleDisplay(w http.ResponseWriter, r *http.Request) {
	macAddress := r.Header.Get("ID")
	accessToken := r.Header.Get("Access-Token")

	if macAddress == "" || accessToken == "" || !validateAPIKey(macAddress, accessToken) {
		s.sendJSONResponse(w, DisplayResponse{
			Status:        500,
			Error:         "Device not found",
			ResetFirmware: true,
		})
		return
	}

	// Get or create device state
	s.stateMutex.Lock()
	if s.deviceState[macAddress] == nil {
		s.deviceState[macAddress] = &DeviceState{CurrentImageIndex: 0} // first image was already returned during setup
	}
	state := s.deviceState[macAddress]
	s.stateMutex.Unlock()

	// Find next image
	nextImage, duration, err := s.getNextImage(state.CurrentImageIndex)
	if err != nil {
		// TODO: return a "no images found" image
		s.sendJSONResponse(w, DisplayResponse{
			Status:      404,
			RefreshRate: DefaultRefreshRate,
		})
		return
	}

	// Update state
	s.stateMutex.Lock()
	state.CurrentImageIndex = nextImage.Index
	state.LastUpdate = time.Now()
	s.stateMutex.Unlock()

	// Build image URL from request
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	host := r.Host
	imageURL := fmt.Sprintf("%s://%s/images/%s", scheme, host, nextImage.Filename)

	s.sendJSONResponse(w, DisplayResponse{
		Status:          0,
		ImageURL:        imageURL,
		Filename:        nextImage.Filename,
		RefreshRate:     duration,
		ResetFirmware:   false,
		UpdateFirmware:  false,
		SpecialFunction: "",
	})
}

type ImageInfo struct {
	Filename string
	Index    int
}

func (s *Server) loadImages() error {
	s.imagesMutex.Lock()
	defer s.imagesMutex.Unlock()

	var images []string
	err := filepath.Walk(s.config.ImageDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Check if it's an image file
		ext := strings.ToLower(filepath.Ext(info.Name()))
		if ext == ".bmp" || ext == ".png" {
			// Get relative path from the image directory
			relPath, err := filepath.Rel(s.config.ImageDir, path)
			if err != nil {
				return err
			}
			images = append(images, relPath)
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Sort images lexicographically
	sort.Strings(images)
	s.images = images

	return nil
}

func (s *Server) getNextImage(currentIndex int) (ImageInfo, int, error) {
	s.imagesMutex.RLock()
	defer s.imagesMutex.RUnlock()

	if len(s.images) == 0 {
		return ImageInfo{}, DefaultRefreshRate, fmt.Errorf("no images found")
	}

	var nextIndex int
	if currentIndex == -1 || currentIndex >= len(s.images)-1 {
		// First request or last image, start from beginning
		nextIndex = 0
	} else {
		// Get next image
		nextIndex = currentIndex + 1
	}

	nextImage := s.images[nextIndex]
	duration := s.parseDurationFromFilename(nextImage)

	return ImageInfo{
		Filename: nextImage,
		Index:    nextIndex,
	}, duration, nil
}

func (s *Server) handleImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]

	if filename == "" {
		http.NotFound(w, r)
		return
	}

	// Construct full path
	filePath := filepath.Join(s.config.ImageDir, filename)

	// Security check: ensure the file is within the image directory
	cleanPath := filepath.Clean(filePath)
	if !strings.HasPrefix(cleanPath, filepath.Clean(s.config.ImageDir)) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Check if file exists
	if _, err := os.Stat(cleanPath); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	// Serve the file
	http.ServeFile(w, r, cleanPath)
}

func (s *Server) handleLog(w http.ResponseWriter, r *http.Request) {
	accessToken := r.Header.Get("Access-Token")
	if accessToken == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var logReq LogRequest
	if err := json.NewDecoder(r.Body).Decode(&logReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Print log to stdout
	logJSON, _ := json.Marshal(logReq.Log)
	log.Printf("Device log [%s]: %s", accessToken[:8], string(logJSON))

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) parseDurationFromFilename(filename string) int {
	// Default duration
	duration := DefaultRefreshRate

	// Check if filename ends with --XX pattern
	base := strings.TrimSuffix(filename, filepath.Ext(filename))
	parts := strings.Split(base, "--")

	if len(parts) >= 2 {
		lastPart := parts[len(parts)-1]
		if d, err := strconv.Atoi(lastPart); err == nil && d > 0 {
			duration = d
		}
	}

	return duration
}

func (s *Server) sendJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
