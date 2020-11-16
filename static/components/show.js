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
  mounted() { window.addEventListener('keyup', this.arrow) },
  beforeUnmount: function () { window.removeEventListener('keyup', this.arrow) },
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
      if ($(window).width() <= 900) $('.sidebar').toggle('slide')
      this.$parent.category = {}
      this.$parent.content = 'category'
    },
    load: function (id, category) {
      if ($(window).width() <= 900) $('.sidebar').toggle('slide')
      this.$parent.content = 'showBookmark'
      this.$parent.current = { id: id, category: category }
    }
  }
}

const showBookmarks = {
  props: {
    current: Object
  },
  data() {
    return {
      bookmark: {
        bookmarks: [],
        category: { name: this.current.category }
      },
      smallSize: window.innerWidth <= 700 ? true : false,
      start: 0
    }
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
          <tr v-for='b in bookmark.bookmarks' :data-id='b.ID'>
            <td>{{ b.Name }}</td>
            <td><a :href='b.URL' target='_blank' class='url' :data-url='b.URL'>{{ b.URL }}</a></td>
            <td>{{ b.Category }}</td>
            <td>
              <a class='icon' @click='edit(b)'><i class='material-icons edit'>edit</i></a>
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>`,
  mounted() {
    this.load(this.$parent.current.id)
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
    current(obj) {
      this.start = 0
      this.load(obj.id)
    },
    smallSize(isSmall) {
      var arr = Array.from(document.getElementsByClassName('url'))
      if (isSmall) arr.forEach(i => i.text = i.text.replace(/https?:\/\/(www\.)?/i, ''))
      else arr.forEach(i => i.text = i.dataset.url)
    }
  },
  methods: {
    load: function (id, more) {
      this.$parent.active = this.$parent.current.id
      this.$parent.loading = true
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
              document.title = this.current.category + ' - My Bookmarks'
            }
          })
        }).then(() => this.$parent.loading = false)
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
          this.load(this.current.id, true)
        }
      }
    },
    editCategory: function () {
      this.$parent.category = {
        Name: this.bookmark.category.name,
        ID: this.bookmark.category.id
      }
      this.$parent.content = 'category'
    },
    add: function () {
      this.$parent.bookmark = {}
      this.$parent.content = 'bookmark'
    },
    edit: function (bookmark) {
      this.$parent.bookmark = bookmark
      this.$parent.content = 'bookmark'
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
