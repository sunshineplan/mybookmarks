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
        <a class='h3 title'>{{ bookmark.category.name }}</a>
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
        <tbody>
          <tr v-for='b in bookmark.bookmarks'>
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
    window.addEventListener('resize', this.checkSize)
    window.addEventListener('scroll', this.checkScroll, true)
  },
  beforeUnmount: function () {
    window.removeEventListener('resize', this.checkSize)
    window.removeEventListener('scroll', this.checkScroll)
  },
  watch: {
    current(obj) { this.load(obj.id) },
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
      post('/bookmark/get', { category: id, start: this.start })
        .then(resp => {
          if (!resp.ok) resp.text().then(err =>
            BootstrapButtons.fire('Error', err, 'error'))
          else resp.json().then(json => {
            if (more)
              this.bookmark.bookmarks = this.bookmark.bookmarks.concat(json.bookmarks)
            else {
              this.bookmark = json
              document.title = this.current.category + ' - My Bookmarks'
            }
            this.$parent.loading = false
          })
        })
        .catch(e => BootstrapButtons.fire('Error', e, 'error'))
        .then(() => this.$parent.loading = false)
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
