package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type LearningPath struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Duration    string    `json:"duration"`
	Difficulty  string    `json:"difficulty"`
	Sections    []Section `json:"sections"`
}

type Section struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Duration    string `json:"duration"`
	Status      string `json:"status"` // "locked", "available", "completed"
}

type ContainerRequest struct {
	SectionID string `json:"sectionId"`
}

type ContainerResponse struct {
	ContainerID string `json:"containerId"`
	Status      string `json:"status"`
}

// Global state management
var (
	containers    = make(map[string]*ContainerInfo)
	containersMux = sync.RWMutex{}
	upgrader      = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow connections from any origin in development
		},
	}
)

type ContainerInfo struct {
	ID        string
	SectionID string
	Status    string
	CreatedAt time.Time
}

type TerminalMessage struct {
	Type string `json:"type"`
	Data string `json:"data"`
}

func main() {
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Routes
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]string{
			"message": "Linux Containerization Learning Platform API",
			"version": "1.0.0",
		})
	})

	// Learning paths
	e.GET("/api/learning-paths", getLearningPaths)
	e.GET("/api/learning-paths/:id", getLearningPath)
	e.GET("/api/learning-paths/:id/sections/:sectionId", getSection)

	// Container management
	e.POST("/api/containers/create", createContainer)
	e.GET("/api/containers/:id", getContainer)
	e.DELETE("/api/containers/:id", deleteContainer)

	// Terminal/Shell endpoints
	e.GET("/api/terminal/:containerId/ws", handleWebSocket)

	log.Println("Server starting on :8080...")
	e.Logger.Fatal(e.Start(":8080"))
}

func getLearningPaths(c echo.Context) error {
	paths := []LearningPath{
		{
			ID:          "linux-containerization",
			Title:       "Linux Containerization Mastery",
			Description: "Complete journey from process management to building your own container runtime",
			Duration:    "3-4 weeks",
			Difficulty:  "Intermediate to Advanced",
			Sections: []Section{
				{
					ID:          "01-process-management",
					Title:       "Process Management",
					Description: "Understanding Linux processes and process isolation",
					Duration:    "1-2 days",
					Status:      "available",
				},
				{
					ID:          "02-namespaces",
					Title:       "Namespaces",
					Description: "Creating isolated environments with Linux namespaces",
					Duration:    "2-3 days",
					Status:      "locked",
				},
				{
					ID:          "03-cgroups",
					Title:       "Control Groups",
					Description: "Resource management with cgroups",
					Duration:    "2-3 days",
					Status:      "locked",
				},
				// Add more sections...
			},
		},
	}

	return c.JSON(http.StatusOK, paths)
}

func getLearningPath(c echo.Context) error {
	id := c.Param("id")

	// For now, return the linux-containerization path
	if id == "linux-containerization" {
		path := LearningPath{
			ID:          "linux-containerization",
			Title:       "Linux Containerization Mastery",
			Description: "Complete journey from process management to building your own container runtime",
			Duration:    "3-4 weeks",
			Difficulty:  "Intermediate to Advanced",
			Sections:    getSections(),
		}
		return c.JSON(http.StatusOK, path)
	}

	return c.JSON(http.StatusNotFound, map[string]string{"error": "Learning path not found"})
}

func getSection(c echo.Context) error {
	// pathId := c.Param("id")
	sectionId := c.Param("sectionId")

	// Mock section data - in real implementation, this would fetch from database
	section := Section{
		ID:          sectionId,
		Title:       "Process Management",
		Description: "Understanding Linux processes and process isolation",
		Duration:    "1-2 days",
		Status:      "available",
	}

	return c.JSON(http.StatusOK, section)
}

func createContainer(c echo.Context) error {
	var req ContainerRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}

	// Generate a mock container ID
	containerID := fmt.Sprintf("mock-container-%d", time.Now().Unix())

	// Store container info
	containersMux.Lock()
	containers[containerID] = &ContainerInfo{
		ID:        containerID,
		SectionID: req.SectionID,
		Status:    "running",
		CreatedAt: time.Now(),
	}
	containersMux.Unlock()

	return c.JSON(http.StatusOK, ContainerResponse{
		ContainerID: containerID,
		Status:      "created",
	})
}

func getContainer(c echo.Context) error {
	containerId := c.Param("id")
	
	containersMux.RLock()
	containerInfo, exists := containers[containerId]
	containersMux.RUnlock()
	
	if !exists {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Container not found",
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"id":        containerInfo.ID,
		"status":    containerInfo.Status,
		"sectionId": containerInfo.SectionID,
		"createdAt": containerInfo.CreatedAt,
	})
}

func deleteContainer(c echo.Context) error {
	containerId := c.Param("id")
	
	// Remove from our tracking
	containersMux.Lock()
	_, exists := containers[containerId]
	if exists {
		delete(containers, containerId)
	}
	containersMux.Unlock()

	if !exists {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Container not found",
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Container " + containerId + " deleted",
	})
}

func handleWebSocket(c echo.Context) error {
	containerId := c.Param("containerId")
	
	// Upgrade HTTP connection to WebSocket
	ws, err := upgrader.Upgrade(c.Response().Writer, c.Request(), nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return err
	}
	defer ws.Close()

	// Check if container exists and is running
	containersMux.RLock()
	containerInfo, exists := containers[containerId]
	containersMux.RUnlock()
	
	if !exists {
		ws.WriteJSON(TerminalMessage{
			Type: "error",
			Data: "Container not found",
		})
		return nil
	}

	// Create a terminal session for this container
	return handleLocalTerminal(ws, containerInfo)
}

func handleLocalTerminal(ws *websocket.Conn, containerInfo *ContainerInfo) error {
	// Create a local bash session with PTY for demonstration
	cmd := exec.Command("/bin/bash")
	
	// Set environment variables
	cmd.Env = append(os.Environ(),
		"TERM=xterm-256color",
		fmt.Sprintf("SECTION_ID=%s", containerInfo.SectionID),
		"PS1=learning-container:$ ",
	)
	
	// Start the command with a pty
	ptmx, err := pty.Start(cmd)
	if err != nil {
		ws.WriteJSON(TerminalMessage{
			Type: "error",
			Data: fmt.Sprintf("Failed to start terminal: %v", err),
		})
		return err
	}
	defer func() {
		ptmx.Close()
		cmd.Process.Kill()
	}()

	// Handle bidirectional communication
	go func() {
		// Read from PTY and send to WebSocket
		buf := make([]byte, 1024)
		for {
			n, err := ptmx.Read(buf)
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Printf("PTY read error: %v", err)
				return
			}
			
			if err := ws.WriteJSON(TerminalMessage{
				Type: "output",
				Data: string(buf[:n]),
			}); err != nil {
				log.Printf("Failed to write to WebSocket: %v", err)
				return
			}
		}
	}()

	// Read from WebSocket and send to PTY
	for {
		var msg TerminalMessage
		if err := ws.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		if msg.Type == "input" {
			// Write to PTY
			if _, err := ptmx.Write([]byte(msg.Data)); err != nil {
				log.Printf("Failed to write to PTY: %v", err)
				break
			}
		} else if msg.Type == "resize" {
			// Handle terminal resize
			var resizeData map[string]interface{}
			if err := json.Unmarshal([]byte(msg.Data), &resizeData); err == nil {
				if cols, ok := resizeData["cols"].(float64); ok {
					if rows, ok := resizeData["rows"].(float64); ok {
						pty.Setsize(ptmx, &pty.Winsize{
							Rows: uint16(rows),
							Cols: uint16(cols),
						})
					}
				}
			}
		}
	}

	return nil
}

func getSections() []Section {
	return []Section{
		{
			ID:          "01-process-management",
			Title:       "Process Management",
			Description: "Understanding Linux processes and process isolation",
			Duration:    "1-2 days",
			Status:      "available",
		},
		{
			ID:          "02-namespaces",
			Title:       "Namespaces",
			Description: "Creating isolated environments with Linux namespaces",
			Duration:    "2-3 days",
			Status:      "locked",
		},
		{
			ID:          "03-cgroups",
			Title:       "Control Groups",
			Description: "Resource management with cgroups",
			Duration:    "2-3 days",
			Status:      "locked",
		},
		{
			ID:          "04-filesystem-isolation",
			Title:       "Filesystem Isolation",
			Description: "Chroot and pivot_root for filesystem isolation",
			Duration:    "1-2 days",
			Status:      "locked",
		},
		{
			ID:          "05-container-images",
			Title:       "Container Images",
			Description: "Understanding layered filesystems and image management",
			Duration:    "2-3 days",
			Status:      "locked",
		},
		{
			ID:          "06-network-virtualization",
			Title:       "Network Virtualization",
			Description: "Virtual networks, bridges, and network namespaces",
			Duration:    "2-3 days",
			Status:      "locked",
		},
		{
			ID:          "07-security-capabilities",
			Title:       "Security & Capabilities",
			Description: "Linux capabilities and container security",
			Duration:    "2 days",
			Status:      "locked",
		},
		{
			ID:          "08-container-runtime",
			Title:       "Container Runtime",
			Description: "OCI specification and runtime implementation",
			Duration:    "2-3 days",
			Status:      "locked",
		},
		{
			ID:          "09-advanced-concepts",
			Title:       "Advanced Concepts",
			Description: "Init systems, signal handling, and process reaping",
			Duration:    "2-3 days",
			Status:      "locked",
		},
		{
			ID:          "10-orchestration-basics",
			Title:       "Orchestration Basics",
			Description: "Multi-container management and service discovery",
			Duration:    "1-2 days",
			Status:      "locked",
		},
	}
}
