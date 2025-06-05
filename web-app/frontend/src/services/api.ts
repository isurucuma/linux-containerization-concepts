const API_BASE_URL = 'http://localhost:8080'

export interface LearningPath {
  id: string
  title: string
  description: string
  duration: string
  difficulty: string
  sections: Section[]
}

export interface Section {
  id: string
  title: string
  description: string
  duration: string
  status: 'locked' | 'available' | 'completed'
}

export interface Container {
  id: string
  status: 'running' | 'stopped' | 'error'
  image: string
}

class ApiService {
  async getLearningPaths(): Promise<LearningPath[]> {
    const response = await fetch(`${API_BASE_URL}/api/learning-paths`)
    if (!response.ok) {
      throw new Error('Failed to fetch learning paths')
    }
    return response.json()
  }

  async getLearningPath(id: string): Promise<LearningPath> {
    const response = await fetch(`${API_BASE_URL}/api/learning-paths/${id}`)
    if (!response.ok) {
      throw new Error('Failed to fetch learning path')
    }
    return response.json()
  }

  async getSection(pathId: string, sectionId: string): Promise<Section> {
    const response = await fetch(`${API_BASE_URL}/api/learning-paths/${pathId}/sections/${sectionId}`)
    if (!response.ok) {
      throw new Error('Failed to fetch section')
    }
    return response.json()
  }

  async createContainer(sectionId: string): Promise<{ containerId: string; status: string }> {
    const response = await fetch(`${API_BASE_URL}/api/containers/create`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        sectionId,
        image: 'linux-containers-env:latest'
      })
    })
    if (!response.ok) {
      throw new Error('Failed to create container')
    }
    return response.json()
  }

  async getContainer(containerId: string): Promise<Container> {
    const response = await fetch(`${API_BASE_URL}/api/containers/${containerId}`)
    if (!response.ok) {
      throw new Error('Failed to get container status')
    }
    return response.json()
  }

  async deleteContainer(containerId: string): Promise<{ message: string }> {
    const response = await fetch(`${API_BASE_URL}/api/containers/${containerId}`, {
      method: 'DELETE'
    })
    if (!response.ok) {
      throw new Error('Failed to delete container')
    }
    return response.json()
  }
}

export const apiService = new ApiService()
