<template>
  <nav class="navbar navbar-light topbar">
    <div class="d-flex" style="height: 100%">
      <a class="toggle" v-if="user && smallSize" @click="toggle">
        <i class="material-icons menu">menu</i>
      </a>
      <router-link class="brand" to="/">My Bookmarks</router-link>
    </div>
    <div class="navbar-nav flex-row" v-if="user">
      <a class="nav-link" v-text="user"></a>
      <a class="nav-link link" @click="setting()">Setting</a>
      <a class="nav-link link" href="/logout">Log Out</a>
    </div>
    <div class="navbar-nav flex-row" v-else>
      <a class="nav-link">Log In</a>
    </div>
  </nav>
  <Login v-if="!user"></Login>
  <div v-else>
    <transition name="slide">
      <Sidebar v-show="showSidebar || !smallSize"></Sidebar>
    </transition>
    <div
      class="content"
      style="padding-left: 250px"
      :style="{ opacity: loading ? 0.5 : 1 }"
      @mousedown="closeSidebar"
      v-if="sidebar"
    >
      <router-view />
    </div>
  </div>
  <div class="loading" v-show="loading">
    <div class="sk-wave sk-center">
      <div class="sk-wave-rect"></div>
      <div class="sk-wave-rect"></div>
      <div class="sk-wave-rect"></div>
      <div class="sk-wave-rect"></div>
      <div class="sk-wave-rect"></div>
    </div>
  </div>
</template>

<script>
import Login from "./components/Login.vue";
import Sidebar from "./components/Sidebar.vue";

export default {
  name: "App",
  components: { Login, Sidebar },
  data() {
    return { smallSize: window.innerWidth <= 900 };
  },
  computed: {
    user() {
      return this.$store.state.username;
    },
    loading() {
      return this.$store.state.loading;
    },
    sidebar() {
      return this.$store.state.sidebar;
    },
    showSidebar() {
      return this.$store.state.showSidebar;
    },
  },
  mounted() {
    window.addEventListener("resize", this.checkSize900);
  },
  beforeUnmount() {
    window.removeEventListener("resize", this.checkSize900);
  },
  methods: {
    checkSize900() {
      this.checkSize(900);
    },
    setting() {
      this.closeSidebar();
      this.$router.push("/setting");
    },
    toggle() {
      this.$store.commit("toggleSidebar");
    },
    closeSidebar() {
      if (this.smallSize) this.$store.commit("closeSidebar");
    },
  },
};
</script>

<style>
:root {
  --sk-color: #1a73e8;
}

body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto,
    "Helvetica Neue", Arial, "Noto Sans", "Microsoft YaHei New",
    "Microsoft Yahei", 微软雅黑, 宋体, SimSun, STXihei, 华文细黑, sans-serif;
}

header {
  height: 100px;
}

.content {
  position: fixed;
  top: 0;
  padding-top: 90px;
  height: 100%;
  width: 100%;
}

.form {
  padding: 0 20px;
}

.form-control {
  width: 250px;
}

.swal {
  margin: 8px 6px;
}

button + button {
  margin-left: 0.3em;
}

.delete {
  margin-top: 8px;
}

.slide-leave-active,
.slide-enter-active {
  transition: 0.5s;
}

.slide-enter-from,
.slide-leave-to {
  transform: translate(-100%, 0);
}

.sortable-ghost {
  opacity: 0;
}

@media (max-width: 900px) {
  .content {
    padding-left: 0 !important;
  }
}
</style>

<style scoped>
.topbar {
  position: fixed;
  top: 0px;
  z-index: 2;
  width: 100%;
  height: 70px;
  padding: 0 10px 0 0;
  background-color: #1a73e8;
  user-select: none;
}

.topbar .nav-link {
  padding-left: 8px;
  padding-right: 8px;
  color: white !important;
}

.topbar .link:hover {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 5px;
  cursor: pointer;
}

.toggle {
  padding: 20px;
  color: white !important;
}

.toggle:hover {
  color: #1a73e8 !important;
  background-color: rgb(232, 232, 232);
}

.menu {
  font-size: 30px;
}

.brand {
  padding-left: 20px;
  margin: auto;
  font-size: 25px;
  letter-spacing: 0.3px;
  color: white;
}

.brand:hover {
  color: white;
  text-decoration: none;
}

.loading {
  position: fixed;
  z-index: 2;
  top: 70px;
  left: 250px;
  height: calc(100% - 70px);
  width: calc(100% - 250px);
  display: flex;
}

@media (max-width: 900px) {
  .brand {
    padding-left: 10px;
  }

  .loading {
    left: 0;
    width: 100%;
  }
}
</style>
