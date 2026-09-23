<template>
	<div class="row">
		<div class="col-xl-6">
			<div class="card">
				<div class="card-body">
					<h4 class="card-title mb-4">Đổi mật khẩu</h4>
					<form @submit.prevent="submitUpdatePassword" method="post">
						<div class="mb-4">
							<label class="form-label required">Mật khẩu cũ</label>
							<input type="password" class="form-control" v-model="oldPassword" @focus="errors.oldPassword = false">
							<ul class="parsley-errors-list filled" v-if="errors.oldPassword">
								<li class="parsley-required">{{ errors.oldPassword }}</li>
							</ul>
						</div>
						<div class="mb-4">
							<label class="form-label required">Mật khẩu mới</label>
							<input type="password" class="form-control" v-model="password" @focus="errors.password = false">
							<ul class="parsley-errors-list filled" v-if="errors.password">
								<li class="parsley-required">{{ errors.password }}</li>
							</ul>
						</div>
						<div class="mb-4">
							<label class="form-label required">Xác nhận mật khẩu mới</label>
							<input type="password" class="form-control" v-model="confirmPassword" @focus="errors.confirmPassword = false">
							<ul class="parsley-errors-list filled" v-if="errors.confirmPassword">
								<li class="parsley-required">{{ errors.confirmPassword }}</li>
							</ul>
						</div>
						<div class="button-items">
							<button type="submit" class="btn btn-success waves-effect btn-label waves-light w-md">
								<i class="bx bx-save label-icon"></i> Cập nhật
							</button>
						</div>
					</form>
				</div>
				<!-- end card body -->
			</div>
			<!-- end card -->
		</div>
	</div>
</template>

<script>
	import admin from '@/api/admin'
	import * as yup from 'yup'
	import {useForm} from 'vee-validate'
	export default {
		name: 'Profile',
		setup() {
			const simpleSchema = yup.object({
				oldPassword: yup.string().required().min(8).max(20),
				password: yup.string().required().min(8).max(20),
				confirmPassword: yup.string().required().min(8).max(20)
			})
			const {useFieldModel, validate, validateField} = useForm({
				initialValues: {
					oldPassword: '',
					password: '',
					confirmPassword: '',
				},
				validationSchema: simpleSchema,
				validateOnMount: false
			})
			const [oldPassword, password, confirmPassword] = useFieldModel(['oldPassword', 'password', 'confirmPassword'])
			return {
				validate,
				validateField,
				oldPassword,
				password,
				confirmPassword
			}
		},
		data: function() {
			return {
				errors: {},
				isSubmitting: false
			}
		},
		methods: {
			async submitUpdatePassword() {
				this.isSubmitting = true
				const {valid, errors} = await this.validate()
				this.errors = errors
				if (valid) {
					let res = await admin.updatePassword({
						oldPassword: this.oldPassword,
						password: this.password,
						confirmPassword: this.confirmPassword
					})
					await this.$store.dispatch('message/setError', res)
					if (res.status === 200) {
						this.oldPassword = ''
						this.password = ''
						this.confirmPassword = ''
					} else if (res.status === 422 && typeof res.data.errors === 'object') {
						this.errors = res.data.errors
					}
				}
				this.isSubmitting = false
			},
		}
	}
</script>

<style scoped>

</style>