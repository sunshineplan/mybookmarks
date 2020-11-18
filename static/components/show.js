const sidebar = {
  computed: {
    active() {
      if (this.$store.state.component != 'showBookmark')
        return false
      return this.$store.state.category.id
    },
    categories() { return this.$store.state.categories },
    total() { return this.categories.reduce((total, i) => total + i.count, 0) }
  },
  template: `
<nav class='nav flex-column navbar-light sidebar'>
  <div class='category-menu'>
    <button class='btn btn-primary btn-sm' @click='add'>Add Category</button>
    <ul class='navbar-nav' v-if='total'>
      <a
        class='navbar-brand category'
        :class='{ active: active == -1 || active == undefined }'
        @click="load(-1, 'All Bookmarks', total)"
      >
        All Bookmarks ({{ total }})
      </a>
      <li v-for='c in categories'>
        <a
          class='nav-link category'
          :class='{ active: active === c.id }'
          @click='load(c.id, c.name, c.count)'
        >
          {{ c.name }} ({{ c.count }})
        </a>
      </li>
    </ul>
  </div>
</nav>`,
  created() {
    this.$store.dispatch('categories')
    this.$store.dispatch('bookmarks', { id: -1 })
  },
  mounted() { window.addEventListener('keyup', this.arrow) },
  beforeUnmount: function () { window.removeEventListener('keyup', this.arrow) },
  methods: {
    arrow: function (event) {
      if (this.active != null) {
        var len = this.categories.length
        var index = this.categories.findIndex(item => item.id == this.active)
        if (event.key == 'ArrowUp') {
          if (index > 0)
            this.load(this.categories[index - 1].id, this.categories[index - 1].name, this.categories[index - 1].count)
          else if (index == 0) this.load(-1, 'All Bookmarks', this.total)
        } else if (event.key == 'ArrowDown')
          if (this.active == -1 && len > 0)
            this.load(this.categories[0].id, this.categories[0].name, this.categories[0].count)
          else if (index >= 0 && index < len - 1)
            this.load(this.categories[index + 1].id, this.categories[index + 1].name, this.categories[index + 1].count)
      }
    },
    add: function () {
      if ($(window).width() <= 900) $('.sidebar').toggle('slide')
      this.$store.commit('editCategory', {})
      this.$store.commit('goto', 'category')
    },
    load: function (id, name, count) {
      if ($(window).width() <= 900) $('.sidebar').toggle('slide')
      this.$store.commit('goto', 'showBookmark')
      if (id != this.active) {
        this.$store.commit('category', { id, name, count, start: 0 })
        this.$store.dispatch('bookmarks', { id })
      }
    }
  }
}

const showBookmarks = {
  data() {
    return { smallSize: window.innerWidth <= 700 ? true : false }
  },
  computed: {
    category() { return this.$store.state.category },
    bookmarks() { return this.$store.state.bookmarks }
  },
  template: `
  <div style='height: 100%'>
    <header style='padding-left: 20px'>
      <div style='height: 50px'>
        <a class='h3'>{{ category.name }}</a>
        <a class='btn icon' v-if='category.id > 0' @click='editCategory'>
          <i class='material-icons edit'>edit</i>
        </a>
      </div>
      <button class='btn btn-primary' @click='add'>Add Bookmark</button>
    </header>
    <div class='table-responsive'>
      <table class='table table-sm'>
        <thead>
          <tr>
            <th>Bookmark</th>
            <th>URL</th>
            <th>Category</th>
            <th></th>
          </tr>
        </thead>
        <tbody id='mybookmarks'>
          <tr v-for='b in bookmarks' :key='b.id' :data-id='b.id'>
            <td>{{ b.name }}</td>
            <td><a :href='b.url' target='_blank' class='url' :data-url='b.url'>{{ b.url }}</a></td>
            <td>{{ b.category }}</td>
            <td>
              <a class='icon' @click='edit(b)'><i class='material-icons edit'>edit</i></a>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>`,
  mounted() {
    document.title = this.category.name + ' - My Bookmarks'
    $('#mybookmarks').sortable(sortable)
    window.addEventListener('resize', this.checkSize)
    window.addEventListener('scroll', this.checkScroll, true)
  },
  beforeUnmount: function () {
    $('#mybookmarks').sortable('destroy')
    window.removeEventListener('resize', this.checkSize)
    window.removeEventListener('scroll', this.checkScroll, true)
  },
  watch: {
    smallSize(isSmall) {
      var arr = Array.from(document.getElementsByClassName('url'))
      if (isSmall) arr.forEach(i => i.text = i.text.replace(/https?:\/\/(www\.)?/i, ''))
      else arr.forEach(i => i.text = i.dataset.url)
    }
  },
  methods: {
    checkSize: function () {
      if (window.innerWidth <= 700) this.smallSize = true
      else this.smallSize = false
    },
    checkScroll: function () {
      var table = document.getElementsByClassName('table-responsive')[0]
      if (table.scrollTop + table.clientHeight >= table.scrollHeight) {
        if (this.category.start + 30 < this.category.count)
          this.$store.dispatch('bookmarks', { more: true })
      }
    },
    editCategory: function () {
      this.$store.commit('editCategory', this.category)
      this.$store.commit('goto', 'category')
    },
    add: function () {
      if (this.category.id > 0)
        this.$store.commit('bookmark', { category: this.category.name })
      else
        this.$store.commit('bookmark', {})
      this.$store.commit('goto', 'bookmark')
    },
    edit: function (bookmark) {
      this.$store.commit('bookmark', bookmark)
      this.$store.commit('goto', 'bookmark')
    }
  }
}

const sortable = {
  update: (event, ui) => {
    var orig = ui.item.data('id'), dest, next
    if (ui.item.prev().length != 0) dest = ui.item.prev().data('id')
    else dest = '#TOP_POSITION#'
    if (ui.item.next().length != 0) next = ui.item.next().data('id')
    else next = '#BOTTOM_POSITION#'
    post('/reorder', { orig, dest, next })
  }
}
