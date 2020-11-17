const sidebar = {
  computed: {
    active() {
      if (this.$store.state.component == 'setting')
        return false
      return this.$store.state.category.id
    },
    categories() { return this.$store.state.categories }
  },
  template: `
<nav class='nav flex-column navbar-light sidebar'>
  <div class='category-menu'>
    <button class='btn btn-primary btn-sm' @click='add'>Add Category</button>
    <ul class='navbar-nav' v-if='categories.total'>
      <a
        class='navbar-brand category'
        :class='{ active: active == -1 || active == undefined }'
        @click="load(-1, 'All Bookmarks')"
      >
        All Bookmarks ({{ categories.total }})
      </a>
      <li v-for='c in categories.categories'>
        <a
          class='nav-link category'
          :class='{ active: active == c.id }'
          @click='load(c.id, c.name)'
        >
          {{ c.name }} ({{ c.count }})
        </a>
      </li>
      <li>
        <a
          class='nav-link category'
          v-if='categories.uncategorized'
          :class='{ active: active === 0 }'
          @click="load(0, 'Uncategorized')"
        >
          Uncategorized ({{ categories.uncategorized }})
        </a>
      </li>
    </ul>
  </div>
</nav>`,
  created() {
    this.$store.commit('loading', true)
    this.$store.commit('categories')
  },
  mounted() { window.addEventListener('keyup', this.arrow) },
  beforeUnmount: function () { window.removeEventListener('keyup', this.arrow) },
  methods: {
    arrow: function (event) {
      if (this.active != null) {
        var len = this.categories.categories.length
        var index = this.categories.categories.findIndex(item => item.id == this.active)
        if (event.key == 'ArrowUp') {
          if (this.active == 0 && len > 0)
            this.load(this.categories.categories[len - 1].id, this.categories.categories[len - 1].name)
          else if (index > 0)
            this.load(this.categories.categories[index - 1].id, this.categories.categories[index - 1].name)
          else if (index == 0) this.load(-1, 'All Bookmarks')
        } else if (event.key == 'ArrowDown')
          if (this.active == -1 && len > 0)
            this.load(this.categories.categories[0].id, this.categories.categories[0].name)
          else if (index >= 0 && index < len - 1)
            this.load(this.categories.categories[index + 1].id, this.categories.categories[index + 1].name)
          else if (index == len - 1) this.load(0, 'Uncategorized')
      }
    },
    add: function () {
      if ($(window).width() <= 900) $('.sidebar').toggle('slide')
      this.$store.commit('editCategory', {})
      this.$store.commit('goto', 'category')
    },
    load: function (id, category) {
      if ($(window).width() <= 900) $('.sidebar').toggle('slide')
      this.$store.commit('goto', 'showBookmark')
      this.$store.commit('category', { id: id, name: category })
    }
  }
}

const showBookmarks = {
  data() {
    return {
      bookmark: {
        bookmarks: [],
        category: { name: this.$store.state.category.name }
      },
      smallSize: window.innerWidth <= 700 ? true : false,
      start: 0
    }
  },
  computed: {
    category() { return this.$store.state.category }
  },
  template: `
  <div style='height: 100%'>
    <header style='padding-left: 20px'>
      <div style='height: 50px'>
        <a class='h3'>{{ bookmark.category.name }}</a>
        <a class='btn icon' v-if='bookmark.category.id > 0' @click='editCategory'>
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
          <tr v-for='b in bookmark.bookmarks' :key='b.id' :data-id='b.id'>
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
    this.load(this.$store.state.category.id)
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
    category(category) {
      this.start = 0
      this.load(category.id)
    },
    smallSize(isSmall) {
      var arr = Array.from(document.getElementsByClassName('url'))
      if (isSmall) arr.forEach(i => i.text = i.text.replace(/https?:\/\/(www\.)?/i, ''))
      else arr.forEach(i => i.text = i.dataset.url)
    }
  },
  methods: {
    load: function (id, more) {
      this.$store.commit('loading', true)
      var params = { category: id }
      if (this.start != 0)
        params.start = this.start
      post('/bookmark/get', params)
        .then(resp => {
          if (!resp.ok) resp.text().then(err => {
            return BootstrapButtons.fire('Error', err, 'error')
          })
          else resp.json().then(json => {
            if (more)
              this.bookmark.bookmarks = this.bookmark.bookmarks.concat(json.bookmarks)
            else {
              this.bookmark = json
              document.title = this.category.name + ' - My Bookmarks'
            }
          })
        }).then(() => this.$store.commit('loading', false))
    },
    checkSize: function () {
      if (window.innerWidth <= 700) this.smallSize = true
      else this.smallSize = false
    },
    checkScroll: function () {
      var table = document.getElementsByClassName('table-responsive')[0]
      if (table.scrollTop + table.clientHeight >= table.scrollHeight) {
        if (this.start + 30 < this.bookmark.total) {
          this.start += 30
          this.load(this.category.id, true)
        }
      }
    },
    editCategory: function () {
      this.$store.commit('editCategory', this.category)
      this.$store.commit('goto', 'category')
    },
    add: function () {
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
    post('/reorder', { orig: orig, dest: dest, next: next })
  }
}
