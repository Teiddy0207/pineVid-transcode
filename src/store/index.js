import {createStore} from 'vuex'
import message from '@/store/modules/message'
import auth from '@/store/modules/auth'

const store = createStore({
	modules: {
		message,
		auth,
	}
})

export default store
