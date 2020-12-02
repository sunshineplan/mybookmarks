<template>
  <div @keyup.enter="setting()">
    <header style="padding-left: 20px">
      <h3>Setting</h3>
      <hr />
    </header>
    <div class="form" :class="{ 'was-validated': validated }">
      <div class="form-group">
        <label for="password">Current Password</label>
        <input
          class="form-control"
          type="password"
          v-model.trim="password"
          id="password"
          maxlength="20"
          required
        />
        <div class="invalid-feedback">This field is required.</div>
      </div>
      <div class="form-group">
        <label for="password1">New Password</label>
        <input
          class="form-control"
          type="password"
          v-model.trim="password1"
          id="password1"
          maxlength="20"
          required
        />
        <div class="invalid-feedback">This field is required.</div>
      </div>
      <div class="form-group">
        <label for="password2">Confirm Password</label>
        <input
          class="form-control"
          type="password"
          v-model.trim="password2"
          id="password2"
          maxlength="20"
          required
        />
        <div class="invalid-feedback">This field is required.</div>
        <small class="form-text text-muted"
          >Max password length: 20 characters.</small
        >
      </div>
      <button class="btn btn-primary" @click="setting()">Change</button>
      <button class="btn btn-primary" @click="goback()">Cancel</button>
    </div>
  </div>
</template>

<script>
import { BootstrapButtons, post, valid } from "../misc.js";

export default {
  name: "Setting",
  data() {
    return {
      password: "",
      password1: "",
      password2: "",
      validated: false,
    };
  },
  mounted() {
    document.title = "Setting - My Bookmarks";
    window.addEventListener("keyup", this.cancel);
  },
  beforeUnmount() {
    window.removeEventListener("keyup", this.cancel);
  },
  methods: {
    async setting() {
      if (valid()) {
        this.validated = false;
        const resp = await post("/setting", {
          password: this.password,
          password1: this.password1,
          password2: this.password2,
        });
        if (!resp.ok)
          await BootstrapButtons.fire("Error", await resp.text(), "error");
        else {
          const json = await resp.json();
          if (json.status == 1) {
            await BootstrapButtons.fire(
              "Success",
              "Your password has changed. Please Re-login!",
              "success"
            );
            this.$store.commit("username", undefined);
            this.$router.push("/");
          } else {
            await BootstrapButtons.fire("Error", json.message, "error");
            if (json.error == 1) this.password = "";
            else {
              this.password1 = "";
              this.password2 = "";
            }
          }
        }
      } else this.validated = true;
    },
  },
};
</script>
