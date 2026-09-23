<template>
	<div class="username-pages my-5 pt-sm-5">
		<div class="container">
			<div class="row justify-content-center">
				<div class="col-md-8 col-lg-6 col-xl-5">
					<div class="card overflow-hidden">
						<div class="bg-primary bg-soft">
							<div class="row">
								<div class="col-7">
									<div class="text-primary p-4">
										<h5 class="text-primary">Chào mừng !</h5>
										<p class="">Đăng nhập Transcode.</p>
									</div>
								</div>
								<div class="col-5 align-self-end">
									<img src="@/assets/images/profile-img.png" class="img-fluid"/>
								</div>
							</div>
						</div>
						<div class="card-body pt-0">
							<div class="auth-logo">
								<a class="auth-logo-light">
									<div class="avatar-md profile-user-wid mb-4">
                    <span class="avatar-title rounded-circle bg-light">
                      <img class="rounded-circle" src="@/assets/logo.svg" height="72"/>
                    </span>
									</div>
								</a>
								<a class="auth-logo-dark">
									<div class="avatar-md profile-user-wid mb-4">
                    <span class="avatar-title rounded-circle bg-light">
                      <img class="rounded-circle" src="@/assets/logo.svg" height="72"/>
                    </span>
									</div>
								</a>
							</div>
							<div class="p-2">
								<form class="form-horizontal" @submit.prevent="login" method="post" autocomplete="off">
									<div class="mb-3">
										<label class="form-label">Tên đăng nhập</label>
										<input class="form-control" type="text" placeholder="Tên đăng nhập"
													 name="username" v-model="username" @focus="errors.username = false"/>
										<ul class="parsley-errors-list filled" v-if="errors.username">
											<li class="parsley-required">{{ errors.username }}</li>
										</ul>
									</div>
									<div class="mb-3">
										<label class="form-label">Mật khẩu</label>
										<div class="input-group auth-pass-inputgroup">
											<input :type="isPasswordShow ? 'text' : 'password'" name="password" class="form-control"
														 placeholder="Mật khẩu" aria-label="Mật khẩu"
														 aria-describedby="password-addon" v-model="password" @focus="errors.password = false"/>
											<button class="btn btn-light" type="button" id="password-addon"
															@click="isPasswordShow = !isPasswordShow">
												<i class="mdi mdi-eye-outline" v-if="!isPasswordShow"></i>
												<i class="mdi mdi-eye-off-outline" v-else></i>
											</button>
										</div>
										<ul class="parsley-errors-list filled" v-if="errors.password">
											<li class="parsley-required">{{ errors.password }}</li>
										</ul>
									</div>
									<div class="form-check">
										<input class="form-check-input" type="checkbox" id="remember-check" v-model="remember">
										<label class="form-check-label" for="remember-check">Ghi nhớ</label>
									</div>
									<div class="mt-3 mb-3 d-grid">
										<button type="submit" class="btn btn-primary waves-effect waves-light" :disabled="isSubmitting">
											<i class="bx bx-hourglass bx-spin font-size-16 align-middle me-2"
												 v-if="isSubmitting"></i>
											Đăng nhập
										</button>
									</div>
								</form>
								<div class="mt-4 text-center">
									<h5 class="font-size-14 mb-3">{{ year }} © Dolphin Technologies</h5>
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>
		</div>
	</div>
</template>

<script>
	import {useForm} from 'vee-validate'
	import * as yup from 'yup'
	import {mapGetters} from 'vuex'

	export default {
		name: 'AuthLogin',
		setup() {
			const simpleSchema = yup.object({
				username: yup.string().required().label('Tên đăng nhập'),
				password: yup.string().required().min(8).label('Mật khẩu')
			})
			const {useFieldModel, validate, validateField} = useForm({
				initialValues: {
					username: '',
					password: '',
					remember: false
				},
				validationSchema: simpleSchema,
				validateOnMount: false,
				initialTouched: {
					username: false,
					password: false
				}
			})
			const [username, password, remember] = useFieldModel(['username', 'password', 'remember'])
			return {
				validate,
				validateField,
				username,
				password,
				remember
			}
		},
		data() {
			return {
				isPasswordShow: false,
				errors: {},
				isSubmitting: false
			}
		},
		computed: {
			...mapGetters('auth', ['getUser']),
			year() {
				return new Date().getFullYear()
			}
		},
		watch: {
			getUser: function () {
				const redirectTo = this.$route.query.redirectTo || {name: 'jobs'}
				this.$router.push(redirectTo)
			}
		},
		methods: {
			async login() {
				try {
					this.isSubmitting = true
					const {valid, errors} = await this.validate()
					this.errors = errors
					if (valid) {
						await this.$store.dispatch('auth/login', {username: this.username, password: this.password, remember: this.remember})
					}
					this.isSubmitting = false
				} catch (error) {
					this.isSubmitting = false
					console.log('login', error)
				}
			}
		}
	}
</script>
