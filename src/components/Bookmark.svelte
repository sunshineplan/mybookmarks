<script lang="ts">
  import { fire, post, valid, confirm } from "../misc";
  import { component, bookmark, categories, bookmarks, total } from "../stores";

  let name = $bookmark ? $bookmark.bookmark : "";
  let url = $bookmark ? $bookmark.url : "";
  let category = $bookmark ? $bookmark.category : "";
  let validated = false;

  $: mode = window.location.pathname == "/bookmark/add" ? "Add" : "Edit";

  const chkURL = () => {
    if (url && !url.match(/^https?:/) && url.length) url = "http://" + url;
  };

  const save = async () => {
    if (valid()) {
      validated = false;
      if (!category) category = "";
      let resp: Response;
      if (mode == "Add")
        resp = await post("/bookmark/add", { bookmark: name, url, category });
      else
        resp = await post("/bookmark/edit/" + $bookmark.id, {
          bookmark: name,
          url,
          category,
        });
      if (!resp.ok) await fire("Error", await resp.text(), "error");
      else {
        const json = await resp.json();
        if (json.status == 1) {
          if (mode == "Add") {
            $bookmarks = [
              ...$bookmarks,
              {
                id: json.id,
                bookmark: name,
                url,
                category,
                seq: $bookmarks.length + 1,
              },
            ];
            const index = $categories.findIndex(
              (c) => c.category === $bookmark.category
            );
            if (index !== -1) $categories[index].count++;
            else if (json.cid)
              $categories.push({ id: json.cid, category, count: 1 });
            $total++;
          } else {
            if (category) {
              const index = $categories.findIndex(
                (c) => c.category === category
              );
              if (index !== -1) $categories[index].count++;
              else $categories.push({ id: json.cid, category, count: 1 });
            }
            if ($bookmark.category)
              $categories[
                $categories.findIndex((c) => c.category === $bookmark.category)
              ].count--;
            const index = $bookmarks.findIndex((b) => b.id === $bookmark.id);
            $bookmarks[index].bookmark = name;
            $bookmarks[index].url = url;
            $bookmarks[index].category = category;
          }
          goback();
        } else {
          await fire("Error", json.message, "error");
          if (json.error == 1) name = "";
          else if (json.error == 2) url = "";
        }
      }
    } else validated = true;
  };

  const del = async () => {
    if (await confirm("bookmark")) {
      const resp = await post("/bookmark/delete/" + $bookmark.id);
      if (!resp.ok) await fire("Error", await resp.text(), "error");
      else {
        const index = $bookmarks.findIndex((b) => b.id === $bookmark.id);
        $bookmarks.splice(index, 1);
        $bookmarks.forEach((b) => {
          if (b.seq > $bookmark.seq) b.seq++;
        });
        if ($bookmark.category)
          $categories[
            $categories.findIndex((c) => c.category === $bookmark.category)
          ].count--;
        $total--;
        goback();
      }
    }
  };

  const goback = () => {
    window.history.pushState({}, "", "/");
    $component = "show";
  };
</script>

<svelte:window
  on:keydown={(e) => {
    if (e.key === "Escape") goback();
  }}
/>

<svelte:head>
  <title>{mode} Bookmark - My Bookmarks</title>
</svelte:head>

<div
  on:keydown={async (e) => {
    if (e.key == "Enter") await save();
  }}
>
  <header style="padding-left: 20px">
    <h3>{mode} Bookmark</h3>
    <hr />
  </header>
  <div class="form" class:was-validated={validated}>
    <div class="form-group">
      <label for="bookmark">Bookmark</label>
      <input
        class="form-control"
        id="bookmark"
        bind:value={name}
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
        id="url"
        type="url"
        bind:value={url}
        on:blur={chkURL}
        required
      />
      <div class="invalid-feedback">Please enter a valid URL.</div>
    </div>
    <div class="form-group">
      <label for="category">Category</label>
      <input
        class="form-control"
        id="category"
        list="category-list"
        bind:value={category}
        maxlength="15"
      />
      <datalist id="category-list">
        {#each $categories as category (category.id)}
          <option>{category.category}</option>
        {/each}
      </datalist>
      <small class="form-text text-muted">
        Max length: 15 characters. One chinese character equal three characters.
      </small>
    </div>
    <button class="btn btn-primary" on:click={save}>{mode}</button>
    <button class="btn btn-primary" on:click={goback}>Cancel</button>
  </div>
  {#if mode == "Edit"}
    <div class="form">
      <button class="btn btn-danger delete" on:click={del}>Delete</button>
    </div>
  {/if}
</div>

<style>
  .delete {
    margin-top: 8px;
  }
</style>
