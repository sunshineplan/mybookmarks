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
      <router-link class="nav-link link" to="/setting">Setting</router-link>
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
</style>
