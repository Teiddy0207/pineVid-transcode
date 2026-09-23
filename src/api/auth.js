import http from './common'

class Auth {
	async login(username, password, remember) {
		return await http.post('/auth/login', {
			username: username,
			password: password,
			remember: remember
		})
	}
	async logout() {
		return await http.post('/auth/logout')
	}
}

export default new Auth()
