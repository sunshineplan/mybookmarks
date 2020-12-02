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
import { BootstrapButtons, post, valid, confirm } from "../utils.js";

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
    save() {
      if (valid()) {
        this.validated = false;
        var r;
        if (this.mode == "Add") r = post("/category/add", { name: this.name });
        else
          r = post("/category/edit/" + this.category.id, { name: this.name });
        r.then((resp) => {
          if (!resp.ok)
            resp
              .text()
              .then((err) => BootstrapButtons.fire("Error", err, "error"));
          else
            resp.json().then((json) => {
              if (json.status == 1) {
                if (this.mode == "Add")
                  this.$store.dispatch("addCategory", this.name);
                else this.$store.dispatch("editCategory", this.name);
                this.goback();
              } else BootstrapButtons.fire("Error", json.message, "error");
            });
        });
      } else this.validated = true;
    },
    del() {
      confirm("category").then((confirm) => {
        if (confirm)
          post("/category/delete/" + this.category.id).then((resp) => {
            if (!resp.ok)
              resp
                .text()
                .then((err) => BootstrapButtons.fire("Error", err, "error"));
            else {
              this.$store.commit("category", {
                id: -1,
                name: "All Bookmarks",
                start: 0,
              });
              this.goback(true);
              this.$store.dispatch("bookmarks", { id: -1 });
            }
          });
      });
    },
  },
};
</script>
