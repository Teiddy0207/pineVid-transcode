<template>
    <top-bar></top-bar>
    <side-bar></side-bar>
    <div class="main-content">
        <div class="page-content">
            <div class="container-fluid">
                <router-view />
            </div>
            <!-- container-fluid -->
        </div>
        <footer class="footer">
            <div class="container-fluid">
                <div class="row">
                    <div class="col-sm-6">{{ year }} © Dolphin Technologies</div>
                </div>
            </div>
        </footer>
    </div>
</template>

<script>
import TopBar from "@/common/partial/TopBar.vue";
import SideBar from "@/common/partial/SideBar.vue";

export default {
    components: { TopBar, SideBar },
    name: "DefaultLayout",
    computed: {
        year() {
            return new Date().getFullYear()
        }
    }
};
</script>