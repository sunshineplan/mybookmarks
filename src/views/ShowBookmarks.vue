<template>
  <div style="height: 100%">
    <header style="padding-left: 20px">
      <div style="height: 50px">
        <a class="h3">{{ category.name }}</a>
        <a class="btn icon" v-if="category.id > 0" @click="editCategory()">
          <i class="material-icons edit">edit</i>
        </a>
      </div>
      <button class="btn btn-primary" @click="add()">Add Bookmark</button>
    </header>
    <div class="table-responsive">
      <table class="table table-sm">
        <thead>
          <tr>
            <th>Bookmark</th>
            <th>URL</th>
            <th>Category</th>
            <th></th>
          </tr>
        </thead>
        <tbody id="mybookmarks">
          <tr v-for="b in bookmarks" :key="b.id">
            <td>{{ b.name }}</td>
            <td>
              <a :href="b.url" target="_blank" class="url" :data-url="b.url">{{
                b.url
              }}</a>
            </td>
            <td>{{ b.category }}</td>
            <td>
              <a class="icon" @click="edit(b)"
                ><i class="material-icons edit">edit</i></a
              >
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>

<script>
import Sortable from "sortablejs";
import { BootstrapButtons, post } from "../utils.js";

export default {
  name: "ShowBookmarks",
  data() {
    return {
      smallSize: window.innerWidth <= 700,
      sortable: "",
    };
  },
  computed: {
    category() {
      return this.$store.state.category;
    },
    bookmarks() {
      return this.$store.state.bookmarks;
    },
  },
  mounted() {
    document.title = this.category.name + " - My Bookmarks";
    this.sortable = new Sortable(document.querySelector("#mybookmarks"), {
      animation: 150,
      delay: 500,
      swapThreshold: 0.5,
      onUpdate: this.onUpdate,
    });
    window.addEventListener("resize", this.checkSize700);
    window.addEventListener("scroll", this.checkScroll, true);
    if (this.smallSize) this.formatURL(true);
  },
  beforeUnmount() {
    this.sortable.destroy();
    window.removeEventListener("resize", this.checkSize700);
    window.removeEventListener("scroll", this.checkScroll, true);
  },
  watch: {
    smallSize(isSmall) {
      this.formatURL(isSmall);
    },
  },
  methods: {
    checkSize700() {
      this.checkSize(700);
    },
    checkScroll() {
      var table = document.querySelector(".table-responsive");
      if (table.scrollTop + table.clientHeight >= table.scrollHeight) {
        if (this.category.start + 30 < this.category.count)
          this.$store.dispatch("bookmarks", { more: true });
      }
    },
    onUpdate(evt) {
      post("/reorder", {
        old: this.bookmarks[evt.oldIndex].id,
        new: this.bookmarks[evt.newIndex].id,
      })
        .then((resp) => resp.text())
        .then((text) => {
          if (text == "1")
            this.$store.dispatch("reorder", {
              old: evt.oldIndex,
              new: evt.newIndex,
            });
          else BootstrapButtons.fire("Error", text, "error");
        });
    },
    formatURL(isSmall) {
      var urls = Array.from(document.querySelectorAll(".url"));
      if (isSmall)
        urls.forEach(
          (url) => (url.text = url.text.replace(/https?:\/\/(www\.)?/i, ""))
        );
      else urls.forEach((url) => (url.text = url.dataset.url));
    },
    editCategory: function () {
      this.$router.push("/category/edit");
    },
    add: function () {
      if (this.category.id <= 0) this.$store.commit("bookmark", {});
      else this.$store.commit("bookmark", { category: this.category.name });
      this.$router.push("/bookmark/add");
    },
    edit: function (bookmark) {
      this.$store.commit("bookmark", bookmark);
      this.$router.push("/bookmark/edit");
    },
  },
};
</script>
