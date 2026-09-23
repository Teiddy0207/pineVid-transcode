import {createApp} from 'vue'
import App from './App.vue'
import router from './router'
import store from './store'
import 'bootstrap'
import axios from 'axios'
import VueAxios from 'vue-axios'
import Notifications from '@kyvg/vue3-notification'
import {setLocale} from 'yup'
import 'vue-datepicker-next/locale/vi'

setLocale({
  mixed: {
    default: 'Dữ liệu không hợp lệ',
    required: 'Dữ liệu không được bỏ trống'
  },
  string: {
    required: 'Dữ liệu không được bỏ trống',
    min: 'Độ dài tối thiểu ${min} ký tự'
  },
  number: {
    typeError: 'Dữ liệu không được bỏ trống',
    required: 'Dữ liệu không được bỏ trống',
    min: 'Giá trị tối thiểu ${min}'
  }
})

const app = createApp(App)
.use(router)
.use(Notifications)
.use(VueAxios, axios)
.use(store)

app.mount('#layout-wrapper')
