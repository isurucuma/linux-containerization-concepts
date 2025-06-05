import { useState, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { Play, BookOpen, Clock, BarChart3 } from 'lucide-react'

interface LearningPath {
  id: string
  title: string
  description: string
  duration: string
  difficulty: string
  sections: Section[]
}

interface Section {
  id: string
  title: string
  description: string
  duration: string
  status: string
}

export default function HomePage() {
  const [learningPaths, setLearningPaths] = useState<LearningPath[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetchLearningPaths()
  }, [])

  const fetchLearningPaths = async () => {
    try {
      const response = await fetch('http://localhost:8080/api/learning-paths')
      const data = await response.json()
      setLearningPaths(data)
    } catch (error) {
      console.error('Failed to fetch learning paths:', error)
    } finally {
      setLoading(false)
    }
  }

  const getDifficultyColor = (difficulty: string) => {
    if (difficulty.toLowerCase().includes('beginner')) return 'text-green-600 bg-green-100'
    if (difficulty.toLowerCase().includes('intermediate')) return 'text-yellow-600 bg-yellow-100'
    if (difficulty.toLowerCase().includes('advanced')) return 'text-red-600 bg-red-100'
    return 'text-blue-600 bg-blue-100'
  }

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-blue-600"></div>
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 via-white to-blue-50">
      {/* Header */}
      <header className="bg-white shadow-sm border-b">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center space-x-3">
              <div className="w-8 h-8 bg-blue-600 rounded-lg flex items-center justify-center">
                <BookOpen className="w-5 h-5 text-white" />
              </div>
              <h1 className="text-xl font-bold text-gray-900">ContainerMaster</h1>
            </div>
            <nav className="flex space-x-8">
              <button className="text-gray-600 hover:text-gray-900">Learning Paths</button>
              <button className="text-gray-600 hover:text-gray-900">Documentation</button>
              <button className="text-gray-600 hover:text-gray-900">Community</button>
            </nav>
          </div>
        </div>
      </header>

      {/* Hero Section */}
      <section className="py-20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <h1 className="text-5xl font-bold text-gray-900 mb-6">
            Master Linux Containerization
          </h1>
          <p className="text-xl text-gray-600 mb-8 max-w-3xl mx-auto">
            Learn container technologies from the ground up with hands-on projects, 
            interactive terminals, and step-by-step guidance. Build your own container runtime!
          </p>
          <div className="flex items-center justify-center space-x-6 text-sm text-gray-500">
            <div className="flex items-center space-x-2">
              <Play className="w-4 h-4" />
              <span>Interactive Learning</span>
            </div>
            <div className="flex items-center space-x-2">
              <BookOpen className="w-4 h-4" />
              <span>Hands-on Projects</span>
            </div>
            <div className="flex items-center space-x-2">
              <BarChart3 className="w-4 h-4" />
              <span>Progressive Difficulty</span>
            </div>
          </div>
        </div>
      </section>

      {/* Learning Paths */}
      <section className="py-16 bg-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-12">
            <h2 className="text-3xl font-bold text-gray-900 mb-4">Learning Paths</h2>
            <p className="text-lg text-gray-600">
              Choose your journey and start building container expertise
            </p>
          </div>

          <div className="grid gap-8 md:grid-cols-1 lg:grid-cols-1">
            {learningPaths.map((path) => (
              <div key={path.id} className="bg-white rounded-xl shadow-lg border border-gray-200 overflow-hidden hover:shadow-xl transition-shadow">
                <div className="p-8">
                  <div className="flex items-start justify-between mb-6">
                    <div className="flex-1">
                      <h3 className="text-2xl font-bold text-gray-900 mb-3">{path.title}</h3>
                      <p className="text-gray-600 mb-4">{path.description}</p>
                      
                      <div className="flex items-center space-x-6 mb-6">
                        <div className="flex items-center space-x-2 text-gray-500">
                          <Clock className="w-4 h-4" />
                          <span className="text-sm">{path.duration}</span>
                        </div>
                        <span className={`px-3 py-1 rounded-full text-xs font-medium ${getDifficultyColor(path.difficulty)}`}>
                          {path.difficulty}
                        </span>
                      </div>
                    </div>
                  </div>

                  {/* Sections Preview */}
                  <div className="mb-6">
                    <h4 className="text-sm font-medium text-gray-900 mb-3">What you'll learn:</h4>
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
                      {path.sections.slice(0, 6).map((section) => (
                        <div key={section.id} className="flex items-center space-x-2 text-sm text-gray-600">
                          <div className="w-2 h-2 bg-blue-600 rounded-full"></div>
                          <span>{section.title}</span>
                        </div>
                      ))}
                      {path.sections.length > 6 && (
                        <div className="text-sm text-gray-500">
                          +{path.sections.length - 6} more sections
                        </div>
                      )}
                    </div>
                  </div>

                  {/* Action Button */}
                  <Link
                    to={`/learning-path/${path.id}`}
                    className="inline-flex items-center px-6 py-3 bg-blue-600 text-white font-medium rounded-lg hover:bg-blue-700 transition-colors"
                  >
                    <Play className="w-4 h-4 mr-2" />
                    Start Learning
                  </Link>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Footer */}
      <footer className="bg-gray-900 text-white py-12">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center">
            <h3 className="text-lg font-semibold mb-4">ContainerMaster</h3>
            <p className="text-gray-400">
              Learn Linux containerization technologies through hands-on experience
            </p>
          </div>
        </div>
      </footer>
    </div>
  )
}
