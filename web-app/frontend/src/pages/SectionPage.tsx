import { useState, useEffect, useRef } from 'react'
import { useParams, Link } from 'react-router-dom'
import { ArrowLeft, Terminal, Play, Square, RefreshCw, BookOpen } from 'lucide-react'
import { Terminal as XTerm } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'

interface Section {
  id: string
  title: string
  description: string
  duration: string
  status: string
}

export default function SectionPage() {
  const { pathId, sectionId } = useParams<{ pathId: string; sectionId: string }>()
  const [section, setSection] = useState<Section | null>(null)
  const [containerId, setContainerId] = useState<string | null>(null)
  const [containerStatus, setContainerStatus] = useState<'stopped' | 'starting' | 'running' | 'error'>('stopped')
  const [loading, setLoading] = useState(true)
  
  const terminalRef = useRef<HTMLDivElement>(null)
  const terminalInstance = useRef<XTerm | null>(null)
  const fitAddon = useRef<FitAddon | null>(null)
  const socketRef = useRef<WebSocket | null>(null)

  useEffect(() => {
    if (pathId && sectionId) {
      fetchSection(pathId, sectionId)
    }
  }, [pathId, sectionId])

  useEffect(() => {
    // Initialize terminal when container is running
    if (containerStatus === 'running' && terminalRef.current && !terminalInstance.current) {
      initializeTerminal()
    }

    return () => {
      // Cleanup terminal and WebSocket
      cleanup()
    }
  }, [containerStatus])

  const cleanup = () => {
    if (socketRef.current) {
      socketRef.current.close()
      socketRef.current = null
    }
    if (terminalInstance.current) {
      terminalInstance.current.dispose()
      terminalInstance.current = null
    }
  }

  const fetchSection = async (pathId: string, sectionId: string) => {
    try {
      const response = await fetch(`http://localhost:8080/api/learning-paths/${pathId}/sections/${sectionId}`)
      const data = await response.json()
      setSection(data)
    } catch (error) {
      console.error('Failed to fetch section:', error)
    } finally {
      setLoading(false)
    }
  }

  const startContainer = async () => {
    setContainerStatus('starting')
    try {
      const response = await fetch('http://localhost:8080/api/containers/create', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          sectionId: sectionId,
          image: 'linux-containers-env:latest'
        })
      })
      
      const data = await response.json()
      setContainerId(data.containerId)
      setContainerStatus('running')
    } catch (error) {
      console.error('Failed to start container:', error)
      setContainerStatus('error')
    }
  }

  const stopContainer = async () => {
    if (!containerId) return
    
    try {
      await fetch(`http://localhost:8080/api/containers/${containerId}`, {
        method: 'DELETE'
      })
      
      setContainerId(null)
      setContainerStatus('stopped')
      
      // Clean up terminal and WebSocket
      cleanup()
    } catch (error) {
      console.error('Failed to stop container:', error)
    }
  }

  const initializeTerminal = () => {
    if (!terminalRef.current || !containerId) return

    // Cleanup existing terminal if any
    if (socketRef.current) {
      socketRef.current.close()
      socketRef.current = null
    }
    if (terminalInstance.current) {
      terminalInstance.current.dispose()
      terminalInstance.current = null
    }

    // Create terminal instance
    const terminal = new XTerm({
      theme: {
        background: '#1e1e1e',
        foreground: '#ffffff',
        cursor: '#ffffff',
      },
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      fontSize: 14,
      rows: 30,
      cols: 120,
    })

    // Create fit addon
    const fit = new FitAddon()
    terminal.loadAddon(fit)

    // Open terminal
    terminal.open(terminalRef.current)
    fit.fit()

    // Store references
    terminalInstance.current = terminal
    fitAddon.current = fit

    // Connect to WebSocket
    const wsUrl = `ws://localhost:8080/api/terminal/${containerId}/ws`
    const socket = new WebSocket(wsUrl)
    socketRef.current = socket

    socket.onopen = () => {
      console.log('WebSocket connected')
      terminal.writeln('Connected to learning environment...')
    }

    socket.onmessage = (event) => {
      try {
        const message = JSON.parse(event.data)
        if (message.type === 'output') {
          terminal.write(message.data)
        } else if (message.type === 'error') {
          terminal.writeln(`\r\n\x1b[31mError: ${message.data}\x1b[0m\r\n`)
        }
      } catch (error) {
        console.error('Failed to parse WebSocket message:', error)
      }
    }

    socket.onerror = (error) => {
      console.error('WebSocket error:', error)
      terminal.writeln('\r\n\x1b[31mConnection error\x1b[0m\r\n')
    }

    socket.onclose = () => {
      console.log('WebSocket disconnected')
      terminal.writeln('\r\n\x1b[33mConnection closed\x1b[0m\r\n')
    }

    // Handle terminal input
    terminal.onData((data) => {
      if (socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({
          type: 'input',
          data: data
        }))
      }
    })

    // Handle terminal resize
    terminal.onResize(({ cols, rows }) => {
      if (socket.readyState === WebSocket.OPEN) {
        socket.send(JSON.stringify({
          type: 'resize',
          data: JSON.stringify({ cols, rows })
        }))
      }
    })

    // Handle window resize
    const handleResize = () => {
      if (fitAddon.current) {
        fitAddon.current.fit()
      }
    }

    window.addEventListener('resize', handleResize)

    // Store cleanup function for later use
    const terminalCleanup = () => {
      window.removeEventListener('resize', handleResize)
      if (socket.readyState === WebSocket.OPEN) {
        socket.close()
      }
      terminal.dispose()
    }

    return terminalCleanup
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-blue-600"></div>
      </div>
    )
  }

  if (!section) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">Section Not Found</h2>
          <Link to={`/learning-path/${pathId}`} className="text-blue-600 hover:text-blue-800">
            Return to Learning Path
          </Link>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center space-x-4">
              <Link 
                to={`/learning-path/${pathId}`}
                className="flex items-center space-x-2 text-gray-600 hover:text-gray-900"
              >
                <ArrowLeft className="w-5 h-5" />
                <span>Back to Learning Path</span>
              </Link>
            </div>
            <div className="flex items-center space-x-3">
              <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center">
                <BookOpen className="w-5 h-5 text-white" />
              </div>
              <h1 className="text-xl font-bold text-gray-900">ContainerMaster</h1>
            </div>
          </div>
        </div>
      </header>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Section Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">{section.title}</h1>
          <p className="text-lg text-gray-600">{section.description}</p>
        </div>

        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* Content Panel */}
          <div className="space-y-6">
            <div className="bg-white rounded-lg shadow-sm border p-6">
              <h2 className="text-xl font-semibold text-gray-900 mb-4">Learning Objectives</h2>
              <ul className="space-y-2 text-gray-600">
                <li>• Understand Linux process management fundamentals</li>
                <li>• Learn about process isolation and namespaces</li>
                <li>• Practice with process creation and monitoring</li>
                <li>• Explore signal handling in containerized environments</li>
              </ul>
            </div>

            <div className="bg-white rounded-lg shadow-sm border p-6">
              <h2 className="text-xl font-semibold text-gray-900 mb-4">Instructions</h2>
              <div className="prose text-gray-600">
                <p>
                  In this section, you'll learn about Linux process management through hands-on exercises. 
                  Use the interactive terminal on the right to explore process concepts.
                </p>
                <p>
                  Start by running the <code>demo</code> command to see process management in action.
                  Then explore the available files and examples.
                </p>
              </div>
            </div>

            <div className="bg-white rounded-lg shadow-sm border p-6">
              <h2 className="text-xl font-semibold text-gray-900 mb-4">Key Concepts</h2>
              <div className="space-y-3">
                <div className="border-l-4 border-blue-500 pl-4">
                  <h3 className="font-semibold text-gray-900">Process Isolation</h3>
                  <p className="text-gray-600 text-sm">How containers isolate processes from the host system</p>
                </div>
                <div className="border-l-4 border-green-500 pl-4">
                  <h3 className="font-semibold text-gray-900">Signal Handling</h3>
                  <p className="text-gray-600 text-sm">Managing process lifecycle through signals</p>
                </div>
                <div className="border-l-4 border-yellow-500 pl-4">
                  <h3 className="font-semibold text-gray-900">Process Trees</h3>
                  <p className="text-gray-600 text-sm">Understanding parent-child process relationships</p>
                </div>
              </div>
            </div>
          </div>

          {/* Terminal Panel */}
          <div className="lg:sticky lg:top-8">
            <div className="bg-white rounded-lg shadow-sm border overflow-hidden">
              {/* Terminal Header */}
              <div className="bg-gray-50 border-b px-4 py-3 flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  <Terminal className="w-5 h-5 text-gray-600" />
                  <h3 className="font-semibold text-gray-900">Interactive Terminal</h3>
                </div>
                <div className="flex items-center space-x-2">
                  {containerStatus === 'stopped' && (
                    <button
                      onClick={startContainer}
                      className="flex items-center space-x-2 px-3 py-1 bg-green-600 text-white text-sm rounded hover:bg-green-700"
                    >
                      <Play className="w-4 h-4" />
                      <span>Start</span>
                    </button>
                  )}
                  {containerStatus === 'starting' && (
                    <div className="flex items-center space-x-2 px-3 py-1 bg-yellow-600 text-white text-sm rounded">
                      <div className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin"></div>
                      <span>Starting...</span>
                    </div>
                  )}
                  {containerStatus === 'running' && (
                    <>
                      <button
                        onClick={() => initializeTerminal()}
                        className="flex items-center space-x-2 px-3 py-1 bg-blue-600 text-white text-sm rounded hover:bg-blue-700"
                      >
                        <RefreshCw className="w-4 h-4" />
                        <span>Reset</span>
                      </button>
                      <button
                        onClick={stopContainer}
                        className="flex items-center space-x-2 px-3 py-1 bg-red-600 text-white text-sm rounded hover:bg-red-700"
                      >
                        <Square className="w-4 h-4" />
                        <span>Stop</span>
                      </button>
                    </>
                  )}
                </div>
              </div>

              {/* Terminal Content */}
              <div className="bg-gray-900">
                {containerStatus === 'stopped' && (
                  <div className="p-8 text-center text-gray-400">
                    <Terminal className="w-12 h-12 mx-auto mb-4 opacity-50" />
                    <p>Click "Start" to launch your interactive learning environment</p>
                  </div>
                )}
                {containerStatus === 'starting' && (
                  <div className="p-8 text-center text-gray-400">
                    <div className="w-8 h-8 border-2 border-gray-400 border-t-transparent rounded-full animate-spin mx-auto mb-4"></div>
                    <p>Setting up your containerized environment...</p>
                  </div>
                )}
                {containerStatus === 'running' && (
                  <div className="terminal-container">
                    <div className="terminal-header">
                      <div className="terminal-dot red"></div>
                      <div className="terminal-dot yellow"></div>
                      <div className="terminal-dot green"></div>
                      <div className="terminal-title">root@container: {section.title}</div>
                    </div>
                    <div 
                      ref={terminalRef} 
                      className="h-96 bg-gray-900"
                    />
                  </div>
                )}
                {containerStatus === 'error' && (
                  <div className="p-8 text-center text-red-400">
                    <p>Failed to start container. Please try again.</p>
                    <button
                      onClick={startContainer}
                      className="mt-4 px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700"
                    >
                      Retry
                    </button>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
