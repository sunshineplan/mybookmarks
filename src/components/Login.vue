<template>
  <div class="content" @keyup.enter="login()">
    <header>
      <h3
        class="d-flex justify-content-center align-items-center"
        style="height: 100%"
      >
        Log In
      </h3>
    </header>
    <div class="login">
      <div class="form-group">
        <label for="username">Username</label>
        <input
          autofocus
          class="form-control"
          v-model.trim="username"
          id="username"
          maxlength="20"
          placeholder="Username"
          required
        />
      </div>
      <div class="form-group">
        <label for="password">Password</label>
        <input
          class="form-control"
          type="password"
          v-model.trim="password"
          id="password"
          maxlength="20"
          placeholder="Password"
          required
        />
      </div>
      <div class="form-group form-check">
        <input
          type="checkbox"
          class="form-check-input"
          v-model="rememberme"
          id="rememberme"
        />
        <label class="form-check-label" for="rememberme">Remember Me</label>
      </div>
      <hr />
      <button class="btn btn-primary login" @click="login()">Log In</button>
    </div>
  </div>
</template>

<script>
import { BootstrapButtons, post } from "../misc.js";

export default {
  name: "Login",
  data() {
    return {
      username: "",
      password: "",
      rememberme: false,
    };
  },
  mounted() {
    document.title = "Log In - My Bookmarks";
  },
  methods: {
    async login() {
      if (!document.querySelector("#username").checkValidity())
        await BootstrapButtons.fire(
          "Error",
          "Username cannot be empty.",
          "error"
        );
      else if (!document.querySelector("#password").checkValidity())
        await BootstrapButtons.fire(
          "Error",
          "Password cannot be empty.",
          "error"
        );
      else {
        const resp = await post("/login", {
          username: this.username,
          password: this.password,
          rememberme: this.rememberme,
        });
        if (!resp.ok)
          await BootstrapButtons.fire("Error", await resp.text(), "error");
        else {
          this.$store.commit("username", this.username);
          this.$router.push("/");
        }
      }
    },
  },
};
</script>

<style scoped>
.login {
  width: 250px;
  margin: 0 auto 20px;
}
</style>
