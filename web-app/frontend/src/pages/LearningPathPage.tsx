import { useState, useEffect } from 'react'
import { useParams, Link } from 'react-router-dom'
import { ArrowLeft, CheckCircle, Lock, Play, Clock, BookOpen } from 'lucide-react'

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
  status: 'locked' | 'available' | 'completed'
}

export default function LearningPathPage() {
  const { pathId } = useParams<{ pathId: string }>()
  const [learningPath, setLearningPath] = useState<LearningPath | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (pathId) {
      fetchLearningPath(pathId)
    }
  }, [pathId])

  const fetchLearningPath = async (id: string) => {
    try {
      const response = await fetch(`http://localhost:8080/api/learning-paths/${id}`)
      const data = await response.json()
      setLearningPath(data)
    } catch (error) {
      console.error('Failed to fetch learning path:', error)
    } finally {
      setLoading(false)
    }
  }

  const getSectionIcon = (status: string) => {
    switch (status) {
      case 'completed':
        return <CheckCircle className="w-5 h-5 text-green-600" />
      case 'available':
        return <Play className="w-5 h-5 text-blue-600" />
      case 'locked':
      default:
        return <Lock className="w-5 h-5 text-gray-400" />
    }
  }

  const getSectionStyle = (status: string) => {
    switch (status) {
      case 'completed':
        return 'bg-green-50 border-green-200 hover:bg-green-100'
      case 'available':
        return 'bg-white border-gray-200 hover:bg-blue-50 cursor-pointer'
      case 'locked':
      default:
        return 'bg-gray-50 border-gray-200 cursor-not-allowed opacity-60'
    }
  }

  const SectionContent = ({ section, index }: { section: Section; index: number }) => (
    <div className="flex items-start space-x-4">
      <div className="flex-shrink-0 flex items-center justify-center w-10 h-10 rounded-full bg-gray-100">
        <span className="text-sm font-semibold text-gray-600">{index + 1}</span>
      </div>
      
      <div className="flex-1 min-w-0">
        <div className="flex items-start justify-between">
          <div className="flex-1">
            <h3 className="text-lg font-semibold text-gray-900 mb-2">
              {section.title}
            </h3>
            <p className="text-gray-600 mb-3">{section.description}</p>
            <div className="flex items-center space-x-4 text-sm text-gray-500">
              <div className="flex items-center space-x-1">
                <Clock className="w-4 h-4" />
                <span>{section.duration}</span>
              </div>
              <span className="capitalize">{section.status}</span>
            </div>
          </div>
          
          <div className="flex-shrink-0 ml-4">
            {getSectionIcon(section.status)}
          </div>
        </div>
      </div>
    </div>
  )

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-blue-600"></div>
      </div>
    )
  }

  if (!learningPath) {
    return (
      <div className="min-h-screen flex items-center justify-center">
        <div className="text-center">
          <h2 className="text-2xl font-bold text-gray-900 mb-4">Learning Path Not Found</h2>
          <Link to="/" className="text-blue-600 hover:text-blue-800">
            Return to Home
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
                to="/" 
                className="flex items-center space-x-2 text-gray-600 hover:text-gray-900"
              >
                <ArrowLeft className="w-5 h-5" />
                <span>Back to Learning Paths</span>
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

      {/* Learning Path Header */}
      <section className="bg-gradient-to-r from-blue-600 to-blue-800 text-white py-16">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="max-w-3xl">
            <h1 className="text-4xl font-bold mb-4">{learningPath.title}</h1>
            <p className="text-xl text-blue-100 mb-6">{learningPath.description}</p>
            <div className="flex items-center space-x-6">
              <div className="flex items-center space-x-2">
                <Clock className="w-5 h-5" />
                <span>{learningPath.duration}</span>
              </div>
              <span className="px-3 py-1 bg-blue-500 rounded-full text-sm font-medium">
                {learningPath.difficulty}
              </span>
            </div>
          </div>
        </div>
      </section>

      {/* Progress Overview */}
      <section className="py-8">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="bg-white rounded-lg shadow-sm border p-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-gray-900">Your Progress</h3>
              <span className="text-sm text-gray-500">
                {learningPath.sections.filter(s => s.status === 'completed').length} of {learningPath.sections.length} sections completed
              </span>
            </div>
            <div className="w-full bg-gray-200 rounded-full h-2">
              <div 
                className="bg-blue-600 h-2 rounded-full transition-all duration-300"
                style={{ 
                  width: `${(learningPath.sections.filter(s => s.status === 'completed').length / learningPath.sections.length) * 100}%` 
                }}
              ></div>
            </div>
          </div>
        </div>
      </section>

      {/* Sections */}
      <section className="py-8">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <h2 className="text-2xl font-bold text-gray-900 mb-8">Learning Sections</h2>
          
          <div className="space-y-4">
            {learningPath.sections.map((section, index) => {
              const isClickable = section.status === 'available' || section.status === 'completed'
              
              return (
                <div key={section.id}>
                  {isClickable ? (
                    <Link
                      to={`/learning-path/${pathId}/section/${section.id}`}
                      className={`block p-6 rounded-lg border transition-all ${getSectionStyle(section.status)}`}
                    >
                      <SectionContent section={section} index={index} />
                    </Link>
                  ) : (
                    <div className={`block p-6 rounded-lg border transition-all ${getSectionStyle(section.status)}`}>
                      <SectionContent section={section} index={index} />
                    </div>
                  )}
                </div>
              )
            })}
          </div>
        </div>
      </section>
    </div>
  )
}
