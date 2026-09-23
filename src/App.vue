<template>
    <router-view/>
    <notifications classes="app-notification"/>
</template>

<script>
    export default {
        name: "App",
        data() {
            return {
                layout: null,
            };
        },
    };
</script>

<style lang="scss">
    @import "sweetalert2/src/sweetalert2.scss";
    @import "vue-datepicker-next/index.css";
    @import "@vueform/multiselect/themes/default.css";
    @import "filepond/dist/filepond.min.css";
    @import "filepond-plugin-image-preview/dist/filepond-plugin-image-preview.min.css";
    @import "datatables.net-bs5";
    @import "./assets/scss/bootstrap.scss";
    @import "./assets/scss/icons.scss";
    @import "./assets/scss/app.scss";
    @import './assets/scss/tooltip.scss';

</style>
