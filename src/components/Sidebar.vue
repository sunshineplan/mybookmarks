<template>
  <nav class="nav flex-column navbar-light sidebar">
    <div class="category-menu">
      <button class="btn btn-primary btn-sm" @click="add()">Add Category</button>
      <ul class="navbar-nav" v-if="total">
        <a
          class="navbar-brand category"
          :class="{ active: active == -1 || active == undefined }"
          @click="load(-1, 'All Bookmarks', total)"
        >
          All Bookmarks ({{ total }})
        </a>
        <li v-for="c in categories" :key="c.id">
          <a
            class="nav-link category"
            :class="{ active: active === c.id }"
            @click="load(c.id, c.name, c.count)"
          >
            {{ c.name }} ({{ c.count }})
          </a>
        </li>
      </ul>
    </div>
  </nav>
</template>

<script>
export default {
  name: "Sidebar",
  computed: {
    active() {
      if (this.$router.currentRoute.value.path != "/") return false;
      return this.$store.state.category.id;
    },
    categories() {
      return this.$store.state.categories;
    },
    total() {
      return this.categories.reduce((total, i) => total + i.count, 0);
    },
  },
  created() {
    this.$store.dispatch("categories");
    this.$store.dispatch("bookmarks", { id: -1 });
  },
  mounted() {
    window.addEventListener("keyup", this.arrow);
  },
  beforeUnmount() {
    window.removeEventListener("keyup", this.arrow);
  },
  methods: {
    arrow(event) {
      if (this.active != null) {
        var len = this.categories.length;
        var index = this.categories.findIndex((item) => item.id == this.active);
        if (event.key == "ArrowUp") {
          if (index > 0)
            this.load(
              this.categories[index - 1].id,
              this.categories[index - 1].name,
              this.categories[index - 1].count
            );
          else if (index == 0) this.load(-1, "All Bookmarks", this.total);
        } else if (event.key == "ArrowDown")
          if (this.active == -1 && len > 0)
            this.load(
              this.categories[0].id,
              this.categories[0].name,
              this.categories[0].count
            );
          else if (index >= 0 && index < len - 1)
            this.load(
              this.categories[index + 1].id,
              this.categories[index + 1].name,
              this.categories[index + 1].count
            );
      }
    },
    add() {
      if (window.innerWidth <= 700) this.$store.commit("closeSidebar");
      this.$router.push("/category/add");
    },
    load(id, name, count) {
      if (window.innerWidth <= 700) this.$store.commit("closeSidebar");
      this.$router.push("/");
      if (id != this.active) {
        this.$store.commit("category", { id, name, count, start: 0 });
        this.$store.dispatch("bookmarks", { id });
      }
    },
  },
};
</script>
