import http from './common'

class Admin {
	async index(options = {}) {
		return await http.get('/admins', options)
	}

	async search(options = {}) {
		return await http.get('/admins/search', options)
	}

	async create(data) {
		return await http.post('/admins', data)
	}

	async read(id) {
		return await http.get(`/admin/${id}`)
	}

	async update(id, data) {
		return await http.patch(`/admin/${id}`, data)
	}

	async delete(id) {
		return await http.delete(`/admin/${id}`)
	}

	async profile() {
		return await http.get('/profile')
	}

	async updateProfile(data) {
		return await http.patch('/profile', data)
	}

	async updatePassword(data) {
		return await http.patch('/profile/password', data)
	}
}

export default new Admin()
