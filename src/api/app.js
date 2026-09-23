import http from './common'

class App {
	async index(options = {}) {
		return http.get('/apps', options)
	}

	async search(options = {}) {
		return await http.get('/apps/search', options)
	}

	async create(data) {
		return await http.post('/apps', data)
	}

	async read(id) {
		return await http.get(`/app/${id}`)
	}

	async update(id, data) {
		return await http.put(`/app/${id}`, data)
	}

	async delete(id) {
		return await http.delete(`/app/${id}`)
	}
}

export default new App()
