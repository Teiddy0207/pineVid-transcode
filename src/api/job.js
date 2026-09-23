import http from './common'

class Job {
	async index(options = {}) {
		return http.get('/jobs', options)
	}

	async search(options = {}) {
		return await http.get('/jobs/search', options)
	}

	async create(data) {
		return await http.post('/jobs', data)
	}

	async read(id) {
		return await http.get(`/job/${id}`)
	}

	async update(id, data) {
		return await http.put(`/job/${id}`, data)
	}

	async delete(id) {
		return await http.delete(`/job/${id}`)
	}
}

export default new Job()
