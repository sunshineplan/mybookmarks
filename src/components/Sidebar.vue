<template>
  <nav class="nav flex-column navbar-light sidebar">
    <div class="category-menu">
      <button class="btn btn-primary btn-sm" @click="add()">
        Add Category
      </button>
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
      return this.$router.currentRoute.value.path != "/"
        ? false
        : this.$store.state.category.id;
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

<style scoped>
.sidebar {
  position: fixed;
  top: 0;
  z-index: 1;
  height: 100%;
  width: 250px;
  padding-top: 70px;
  user-select: none;
}

.category-menu {
  height: 100%;
  width: 100%;
  padding-top: 10px;
  overflow-x: hidden;
  border-right: 1px solid #e9ecef;
  background-color: white;
}

.category-menu .btn {
  margin-left: 20px;
  margin-bottom: 5px;
}

.category-menu .navbar-brand {
  text-indent: 10px;
}

.category-menu .navbar-nav {
  text-indent: 20px;
}

.category-menu .nav-link:hover {
  background-color: rgb(232, 232, 232);
}

#categories {
  height: calc(100% - 76px);
  overflow-y: auto;
}

.category {
  display: block;
  cursor: pointer;
  margin: 0;
  border-left: 5px solid transparent;
  color: rgba(0, 0, 0, 0.7) !important;
}

.active {
  border-left: 5px solid #1a73e8;
  color: #1a73e8 !important;
}

.nav-link.active {
  background-color: #eaf5fd;
}

@media (min-width: 901px) {
  .sidebar {
    display: block !important;
  }
}

@media (max-width: 900px) {
  .sidebar {
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
  }
}
</style>
