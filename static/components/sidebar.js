const sidebar = {
  props: { active: Number },
  data() { return { category: {} } },
  template: `
<nav class='nav flex-column navbar-light sidebar'>
  <div class='category-menu'>
    <button class='btn btn-primary btn-sm' @click='add'>Add Category</button>
    <ul class='navbar-nav' v-if='category.total'>
      <a
        class='navbar-brand category'
        :class='{ active: active == -1 }'
        @click="load(-1, 'All Bookmarks')"
      >
        All Bookmarks ({{ category.total }})
      </a>
      <li v-for='c in category.categories'>
        <a
          class='nav-link category'
          :class='{ active: active == c.ID }'
          @click='load(c.ID, c.Name)'
        >
          {{ c.Name }} ({{ c.Count }})
        </a>
      </li>
      <li>
        <a
          class='nav-link category'
          v-if='category.uncategorized'
          :class='{ active: active == 0 }'
          @click="load(0, 'Uncategorized')"
        >
          Uncategorized ({{ category.uncategorized }})
        </a>
      </li>
    </ul>
  </div>
</nav>`,
  created() {
    this.$parent.loading = true
    post('/category/get')
      .then(response => response.json())
      .then(json => {
        this.category = json
        this.$parent.siderbar = true
        this.$parent.loading = false
      })
  },
  mounted() {
    window.addEventListener('keyup', this.arrow)
  },
  beforeUnmount: function () {
    window.removeEventListener('keyup', this.arrow)
  },
  methods: {
    arrow: function (event) {
      if (this.active != null) {
        var len = this.category.categories.length
        var index = this.category.categories.findIndex(item => item.ID == this.active)
        if (event.key == 'ArrowUp') {
          if (this.active == 0 && len > 0)
            this.load(this.category.categories[len - 1].ID, this.category.categories[len - 1].Name)
          else if (index > 0)
            this.load(this.category.categories[index - 1].ID, this.category.categories[index - 1].Name)
          else if (index == 0) this.load(-1, 'All Bookmarks')
        } else if (event.key == 'ArrowDown')
          if (this.active == -1 && len > 0)
            this.load(this.category.categories[0].ID, this.category.categories[0].Name)
          else if (index >= 0 && index < len - 1)
            this.load(this.category.categories[index + 1].ID, this.category.categories[index + 1].Name)
          else if (index == len - 1) this.load(0, 'Uncategorized')
      }

    },
    add: function () {
      this.$parent.category = {}
      this.$parent.content = 'category'
    },
    load: function (id, category) {
      this.$parent.content = 'showBookmark'
      this.$parent.current = { id: id, category: category }
    }
  }
}
