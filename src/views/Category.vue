<template>
  <div @keyup.enter="save()">
    <header style="padding-left: 20px">
      <h3>{{ mode }} Category</h3>
      <hr />
    </header>
    <div class="form" :class="{ 'was-validated': validated }">
      <div class="form-group">
        <label for="category">Category</label>
        <input
          class="form-control"
          v-model.trim="name"
          id="category"
          maxlength="15"
          required
        />
        <div class="invalid-feedback">This field is required.</div>
        <small class="form-text text-muted">
          Max length: 15 characters. One chinese character equal three
          characters.
        </small>
      </div>
      <button class="btn btn-primary" @click="save()">{{ mode }}</button>
      <button class="btn btn-primary" @click="goback()">Cancel</button>
    </div>
    <div class="form" v-if="mode == 'Edit'">
      <button class="btn btn-danger delete" @click="del()">Delete</button>
    </div>
  </div>
</template>

<script>
import { BootstrapButtons, post, valid, confirm } from "../misc.js";

export default {
  name: "Category",
  data() {
    return {
      name: "",
      validated: false,
    };
  },
  computed: {
    category() {
      return this.$route.params.mode == "edit"
        ? this.$store.state.category
        : {};
    },
    mode() {
      return this.$route.params.mode == "add" ? "Add" : "Edit";
    },
  },
  created() {
    this.name = this.category.name;
  },
  mounted() {
    document.title = this.mode + " Category - My Bookmarks";
    window.addEventListener("keyup", this.cancel);
  },
  beforeUnmount: function () {
    window.removeEventListener("keyup", this.cancel);
  },
  watch: {
    category(category) {
      this.name = category.name;
      document.title = this.mode + " Category - My Bookmarks";
    },
  },
  methods: {
    async save() {
      if (valid()) {
        this.validated = false;
        let resp;
        if (this.mode == "Add")
          resp = await post("/category/add", { name: this.name });
        else
          resp = await post("/category/edit/" + this.category.id, {
            name: this.name,
          });
        if (!resp.ok)
          await BootstrapButtons.fire("Error", await resp.text(), "error");
        else {
          const json = await resp.json();
          if (json.status == 1) {
            if (this.mode == "Add")
              await this.$store.dispatch("addCategory", this.name);
            else await this.$store.dispatch("editCategory", this.name);
            this.goback();
          } else await BootstrapButtons.fire("Error", json.message, "error");
        }
      } else this.validated = true;
    },
    async del() {
      if (await confirm("category")) {
        const resp = await post("/category/delete/" + this.category.id);
        if (!resp.ok)
          await BootstrapButtons.fire("Error", await resp.text(), "error");
        else {
          this.$store.commit("category", {
            id: -1,
            name: "All Bookmarks",
            start: 0,
          });
          await this.goback(true);
          await this.$store.dispatch("bookmarks", { id: -1 });
        }
      }
    },
  },
};
</script>
