<template>
	<div class="row">
		<div class="col-12">
			<div class="page-title-box d-sm-flex align-items-center justify-content-between">
				<h4 class="mb-sm-0 font-size-18">Quản lý ứng dụng</h4>
				<div class="page-title-right">
					<a href="javascript:void(0)" @click="elements.appModalCreate.show()"
						 class="btn btn-success waves-effect btn-label waves-light">
						<i class="bx bx-plus label-icon"></i>
						Thêm mới
					</a>
				</div>
			</div>
		</div>
		<div class="col-lg-12">
			<div class="card">
				<div class="card-body">
					<div class="row mb-3">
						<div class="col-12 col-lg-5">
							<label>Tìm kiếm</label>
							<input type="text" class="form-control" v-model="query.search" @change="forceRender"
										 placeholder="Tìm kiếm ..."/>
						</div>
						<div class="col-12 col-lg-3">
							<label>Trạng thái</label>
							<Multiselect v-model="query.status" :options="options.status" noOptionsText="Không có lựa chọn"
													 noResultsText="Không có kết quả" :searchable="false" :canDeselect="true" :canClear="true"
													 @select="forceRender" @clear="forceRender"/>
						</div>
					</div>

					<data-table class="table table-bordered dt-responsive nowrap w-100" :options="dataTableConfigs"
											ref="appTable">
					</data-table>
				</div>
			</div>
		</div>
	</div>
	<div class="modal fade" tabindex="-1" role="dialog" aria-labelledby="app-modal-create-label" aria-hidden="true"
			 data-bs-backdrop="static" id="app-modal-create">
		<div class="modal-dialog">
			<div class="modal-content">
				<div class="modal-header">
					<h5 class="modal-title" id="app-modal-create-label">Thêm ứng dụng</h5>
				</div>
				<div class="modal-body">
					<form @submit.prevent="submitCreate" method="post">
						<div class="row">
							<div class="col-12">
								<div class="mb-3">
									<label class="form-label required">Tên ứng dụng</label>
									<input type="text" class="form-control" v-model="name" @focus="errors.name = false">
									<ul class="parsley-errors-list filled" v-if="errors.name">
										<li class="parsley-required">{{ errors.name }}</li>
									</ul>
								</div>
							</div>
							<div class="col-12">
								<div class="mb-3">
									<label class="form-label required">Định danh</label>
									<input type="text" class="form-control" v-model="code" @focus="errors.code = false">
									<ul class="parsley-errors-list filled" v-if="errors.code">
										<li class="parsley-required">{{ errors.code }}</li>
									</ul>
								</div>
							</div>
							<div class="col-12">
								<div class="mb-3">
									<label class="form-label required">Mã xác thực</label>
									<div class="input-group">
										<input type="text" class="form-control" placeholder="Mã xác thực" aria-label="Mã xác thực"
													 aria-describedby="key-addon" v-model="key" @focus="errors.key = false">
										<button class="btn btn-success" type="button" id="key-addon" title="Tạo ngẫu nhiên"
														@click="randomKey">
											<i class="mdi mdi-shuffle"></i>
										</button>
									</div>
									<ul class="parsley-errors-list filled" v-if="errors.key">
										<li class="parsley-required">{{ errors.key }}</li>
									</ul>
								</div>
							</div>
							<div class="col-12">
								<div class="mb-3">
									<label>Trạng thái</label>
									<Multiselect v-model="status" :options="options.status" noOptionsText="Không có lựa chọn"
															 noResultsText="Không có kết quả" :searchable="false" :canDeselect="false"
															 :canClear="false"/>
								</div>
							</div>
						</div>
						<div class="button-items">
							<button type="submit" class="btn btn-success waves-effect btn-label waves-light w-md"><i
											class="bx bx-save label-icon"></i>Lưu
							</button>
							<button type="button" class="btn btn-light waves-effect waves-light w-md"
											@click="closeModal">Đóng
							</button>
						</div>
					</form>
				</div>
			</div>
		</div>
	</div>
	<div class="modal fade" tabindex="-1" role="dialog" aria-labelledby="app-modal-update-label" aria-hidden="true"
			 data-bs-backdrop="static" id="app-modal-update">
		<div class="modal-dialog">
			<div class="modal-content">
				<div class="modal-header">
					<h5 class="modal-title" id="app-modal-update-label">Cập nhật ứng dụng</h5>
				</div>
				<div class="modal-body">
					<div class="spinner-border text-primary m-1" role="status" v-if="isModalLoading">
						<span class="sr-only">Đang tải ...</span>
					</div>
					<form @submit.prevent="submitUpdate" method="post" v-else>
						<div class="row">
							<div class="col-12">
								<div class="mb-3">
									<label class="form-label required">Tên ứng dụng</label>
									<input type="text" class="form-control" v-model="name" @focus="errors.name = false">
									<ul class="parsley-errors-list filled" v-if="errors.name">
										<li class="parsley-required">{{ errors.name }}</li>
									</ul>
								</div>
							</div>
							<div class="col-12">
								<div class="mb-3">
									<label class="form-label required">Định danh</label>
									<input type="text" class="form-control" v-model="code" @focus="errors.code = false">
									<ul class="parsley-errors-list filled" v-if="errors.code">
										<li class="parsley-required">{{ errors.code }}</li>
									</ul>
								</div>
							</div>
							<div class="col-12">
								<div class="mb-3">
									<label class="form-label required">Mã xác thực</label>
									<div class="input-group">
										<input type="text" class="form-control" placeholder="Mã xác thực" aria-label="Mã xác thực"
													 aria-describedby="key-addon-update" v-model="key" @focus="errors.key = false">
										<button class="btn btn-success" type="button" id="key-addon-update" title="Tạo ngẫu nhiên"
														@click="randomKey">
											<i class="mdi mdi-shuffle"></i>
										</button>
									</div>
									<ul class="parsley-errors-list filled" v-if="errors.key">
										<li class="parsley-required">{{ errors.key }}</li>
									</ul>
								</div>
							</div>
							<div class="col-12">
								<div class="mb-3">
									<label>Trạng thái</label>
									<Multiselect v-model="status" :options="options.status" noOptionsText="Không có lựa chọn"
															 noResultsText="Không có kết quả" :searchable="false" :canDeselect="false"
															 :canClear="false"/>
								</div>
							</div>
						</div>
						<div class="button-items">
							<button type="submit" class="btn btn-success waves-effect btn-label waves-light w-md"><i
											class="bx bx-save label-icon"></i>Lưu
							</button>
							<button type="button" class="btn btn-light waves-effect waves-light w-md"
											@click="closeModal">Đóng
							</button>
						</div>
					</form>
				</div>
			</div>
		</div>
	</div>
</template>

<script>
	import {Modal} from 'bootstrap'
	import app from '@/api/app'
	import * as yup from 'yup'
	import {useForm} from 'vee-validate'
	import Swal from 'sweetalert2'
	import {mapGetters} from 'vuex'
	import Multiselect from '@vueform/multiselect'
	import {baseConfig, DataTable} from '../../plugins/datatable'
	import {h, render} from 'vue'

	export default {
		name: 'AppIndex',
		components: {
			DataTable,
			Multiselect
		},
		setup() {
			const simpleSchema = yup.object({
				name: yup.string().required(),
				code: yup.string().required(),
				key: yup.string().required(),
				status: yup.string().required()
			})
			const {useFieldModel, validate, validateField} = useForm({
				initialValues: {
					name: '',
					code: '',
					key: '',
					status: 'active'
				},
				validationSchema: simpleSchema,
				validateOnMount: false
			})
			const [name, code, key, status] = useFieldModel(['name', 'code', 'key', 'status'])
			return {
				validate,
				validateField,
				name,
				code,
				key,
				status
			}
		},
		data: function () {
			return {
				app: {},
				dataTableConfigs: {
					...baseConfig,
					...{
						ajax: {
							url: '/api/backend/apps',
							data: this.tableQuery
						},
						columns: [
							{
								className: 'align-middle',
								data: 'name',
								name: 'name',
								orderable: true,
								title: 'Tên ứng dụng',
								render: function (data, type, row, meta) {
									return `<p>${row.name}</p><p class="mb-0">Định danh: <span class="badge rounded-pill bg-success font-size-12">${row.code}</span></p>`
								}
							},
							{
								className: 'align-middle text-center',
								data: 'status',
								name: 'status',
								orderable: false,
								title: 'Trạng thái',
								width: '120px',
								render: function (data, type, row, meta) {
									if (data === 'active') {
										return '<span class="badge rounded-pill bg-success font-size-12">Kích hoạt</span>'
									} else {
										return '<span class="badge rounded-pill bg-danger font-size-12">Vô hiệu</span>'
									}
								}
							},
							{
								className: 'text-center align-middle',
								data: 'id',
								title: 'Hành động',
								width: '120px',
								orderable: false,
								render: function () {
									return ''
								}
							}
						],
						ordering: [0, 'desc'],
						serverSide: true,
						rowCallback: this.rowCallback
					}
				},
				elements: {
					appTable: null,
					appModalCreate: null,
					appModalUpdate: null
				},
				errors: {},
				isSubmitting: false,
				isModalLoading: false,
				options: {
					status: [{value: 'active', label: 'Kích hoạt'}, {value: 'disable', label: 'Vô hiệu'}]
				},
				query: {
					search: '',
					status: ''
				}
			}
		},
		computed: {
			...mapGetters('auth', ['getUser'])
		},
		watch: {},
		methods: {
			tableQuery(d) {
				let order = ''
				let dir = 'asc'
				if (d.order && d.order.length > 0) {
					order = d.columns[d.order[0].column].name
					dir = d.order[0].dir
				}
				return {
					draw: d.draw,
					limit: d.length,
					page: d.start / d.length,
					search: this.query.search,
					status: this.query.status,
					order: order,
					dir: dir
				}
			},
			forceRender() {
				this.$nextTick(function () {
					this.$refs.appTable.dt().ajax.reload(null, false)
				})
			},
			resetForm() {
				this.name = ''
				this.code = ''
				this.key = ''
				this.status = 'active'
			},
			closeModal() {
				this.resetForm()
				this.elements.appModalCreate?.hide()
				this.elements.appModalUpdate?.hide()
			},
			rowCallback(row, data) {
				let group = h('div', {class: 'button-items'}, [])
				let update = h('button', {
						class: 'btn btn-outline-primary waves-effect waves-light btn-sm',
						onClick: () => {
							this.updateApp(data.id)
						}
					}, [h('i', {class: 'bx bx-pencil font-size-16'})]
				)
				let remove = h('button', {
						class: 'btn btn-outline-danger waves-effect waves-light btn-sm',
						onClick: () => {
							this.submitDelete(data.id)
						}
					}, [h('i', {class: 'bx bx-trash font-size-16'})]
				)
				group.children.push(update)
				group.children.push(remove)
				render(group, row.lastElementChild)
			},
			randomKey() {
				let result = ''
				let characters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
				let charactersLength = characters.length
				for (let i = 0; i < 32; i++) {
					result += characters.charAt(Math.floor(Math.random() * charactersLength))
				}
				this.key = result
			},
			async updateApp(id) {
				this.app.id = id
				this.elements.appModalUpdate?.show()
				this.isModalLoading = true
				this.errors = {}
				let res = await app.read(id)
				if (res.status === 200) {
					this.name = res.data.name
					this.code = res.data.code
					this.key = res.data.key
					this.status = res.data.status
				} else {
					this.elements.appModalUpdate?.hide()
					await this.$store.dispatch('message/setError', res)
				}
				this.isModalLoading = false
				await this.$store.dispatch('app/loadApps')
			},
			async submitCreate() {
				this.isSubmitting = true
				const {valid, errors} = await this.validate()
				this.errors = errors
				if (valid) {
					let res = await app.create({
						name: this.name,
						code: this.code,
						key: this.key,
						status: this.status
					})
					await this.$store.dispatch('message/setError', res)
					if (res.status === 201) {
						this.elements.appModalCreate?.hide()
						this.forceRender()
						this.resetForm()
					} else if (res.status === 422 && typeof res.data.errors === 'object') {
						this.errors = res.data.errors
					}
				}
				this.isSubmitting = false
				await this.$store.dispatch('app/loadApps')
			},
			async submitUpdate() {
				this.isSubmitting = true
				const {valid, errors} = await this.validate()
				this.errors = errors
				if (valid) {
					let res = await app.update(this.app.id, {
						name: this.name,
						code: this.code,
						key: this.key,
						status: this.status
					})
					await this.$store.dispatch('message/setError', res)
					if (res.status === 200) {
						this.elements.appModalUpdate?.hide()
						this.forceRender()
						this.resetForm()
					} else if (res.status === 422 && typeof res.data.errors === 'object') {
						this.errors = res.data.errors
					}
				}
				this.isSubmitting = false
				await this.$store.dispatch('app/loadApps')
			},
			async submitDelete(id) {
				let confirm = await Swal.fire({
					title: 'Cảnh báo',
					text: 'Bạn có chắc muốn xóa?',
					icon: 'question',
					confirmButtonText: 'Có, chắc chắn',
					cancelButtonText: 'Không',
					confirmButtonColor: '#f46a6a',
					customClass: {
						closeButton: 'btn btn-primary waves-effect waves-light',
						confirmButton: 'btn btn-danger waves-effect waves-light',
						cancelButton: 'btn btn-primary waves-effect waves-light'
					},
					showConfirmButton: true,
					showCancelButton: true,
					focusConfirm: false,
					focusCancel: true,
					showLoaderOnConfirm: true,
					preConfirm: () => {
					}
				})
				if (confirm.isConfirmed) {
					let res = await app.delete(id)
					this.forceRender()
					await this.$store.dispatch('message/setError', res)
				}
				await this.$store.dispatch('app/loadApps')
			}
		},
		mounted() {
			this.$nextTick(function () {
				this.elements.appModalCreate = new Modal(document.getElementById('app-modal-create'))
				this.elements.appModalUpdate = new Modal(document.getElementById('app-modal-update'))
			})
		},
		unmounted() {
		}
	}
</script>
