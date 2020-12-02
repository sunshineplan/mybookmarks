<template>
  <div @keyup.enter="save()">
    <header style="padding-left: 20px">
      <h3>{{ mode }} Bookmark</h3>
      <hr />
    </header>
    <div class="form" :class="{ 'was-validated': validated }">
      <div class="form-group">
        <label for="bookmark">Bookmark</label>
        <input
          class="form-control"
          v-model.trim="name"
          id="bookmark"
          maxlength="40"
          required
        />
        <div class="invalid-feedback">This field is required.</div>
        <small class="form-text text-muted">Max length: 40 characters.</small>
      </div>
      <div class="form-group">
        <label for="url">URL</label>
        <input
          class="form-control"
          type="url"
          v-model.trim="url"
          id="url"
          @blur="chkURL"
          required
        />
        <div class="invalid-feedback">Please enter a valid URL.</div>
      </div>
      <div class="form-group">
        <label for="category">Category</label>
        <input
          class="form-control"
          list="category-list"
          v-model.trim="category"
          id="category"
          maxlength="15"
        />
        <datalist id="category-list">
          <option v-for="c in categories" :key="c.id">{{ c.name }}</option>
        </datalist>
        <small class="form-text text-muted"
          >Max length: 15 characters. One chinese character equal three
          characters.</small
        >
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
  name: "Bookmark",
  data() {
    return {
      categories: this.$store.state.categories,
      bookmark: this.$store.state.bookmark,
      name: "",
      url: "",
      category: "",
      validated: false,
    };
  },
  computed: {
    mode() {
      return this.$route.params.mode == "add" ? "Add" : "Edit";
    },
  },
  created() {
    this.name = this.bookmark.name;
    this.url = this.bookmark.url;
    this.category = this.bookmark.category;
  },
  mounted() {
    document.title = this.mode + " Bookmark - My Bookmarks";
    window.addEventListener("keyup", this.cancel);
  },
  beforeUnmount: function () {
    window.removeEventListener("keyup", this.cancel);
  },
  methods: {
    chkURL() {
      if (this.url && !this.url.match(/^https?:/) && this.url.length)
        this.url = "http://" + this.url;
    },
    save() {
      if (valid()) {
        this.validated = false;
        var r;
        if (this.mode == "Add")
          r = post("/bookmark/add", {
            name: this.name,
            url: this.url,
            category: this.category,
          });
        else
          r = post("/bookmark/edit/" + this.bookmark.id, {
            name: this.name,
            url: this.url,
            category: this.category,
          });
        r.then((resp) => {
          if (!resp.ok)
            resp
              .text()
              .then((err) => BootstrapButtons.fire("Error", err, "error"));
          else
            resp.json().then((json) => {
              if (json.status == 1) {
                this.goback(true);
                this.$store.dispatch("bookmarks", {
                  id: this.$store.state.category.id,
                });
              } else
                BootstrapButtons.fire("Error", json.message, "error").then(
                  () => {
                    if (json.error == 1) this.name = "";
                    else if (json.error == 2) this.url = "";
                  }
                );
            });
        });
      } else this.validated = true;
    },
    del() {
      confirm("bookmark").then((confirm) => {
        if (confirm)
          post("/bookmark/delete/" + this.bookmark.id).then((resp) => {
            if (!resp.ok)
              resp
                .text()
                .then((err) => BootstrapButtons.fire("Error", err, "error"));
            else {
              this.goback();
              this.$store.dispatch("delBookmarks", this.bookmark);
            }
          });
      });
    },
  },
};
</script>
