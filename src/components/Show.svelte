<script lang="ts">
  import Sortable from "sortablejs";
  import { onMount, createEventDispatcher } from "svelte";
  import { fire, post, confirm } from "../misc";
  import {
    component,
    loading,
    bookmark,
    bookmarks,
    category,
    categories,
  } from "../stores";
  import type { Bookmark } from "../stores";

  const dispatch = createEventDispatcher();
  const isSmall = 700;

  let smallSize = window.innerWidth <= isSmall;
  let editable = false;

  $: currentBookmarks =
    $category.id === -1
      ? $bookmarks
      : $bookmarks.filter(
          (bookmark) => bookmark.category === $category.category
        );

  onMount(() => {
    const sortable = new Sortable(
      document.querySelector("#mybookmarks") as HTMLElement,
      {
        animation: 150,
        delay: 500,
        swapThreshold: 0.5,
        onUpdate,
      }
    );
    if (smallSize) formatURL(true);
    return () => sortable.destroy();
  });

  const onUpdate = async (evt: Sortable.SortableEvent) => {
    const resp = await post("/reorder", {
      old: currentBookmarks[evt.oldIndex as number].id,
      new: currentBookmarks[evt.newIndex as number].id,
    });
    if ((await resp.text()) == "1") {
      const current = currentBookmarks[evt.oldIndex as number].id;
      const oldSeq = currentBookmarks[evt.oldIndex as number].seq;
      const newSeq = currentBookmarks[evt.newIndex as number].seq;
      if (oldSeq > newSeq)
        $bookmarks.forEach((b) => {
          if (b.seq >= newSeq && b.seq < oldSeq) b.seq++;
        });
      else
        $bookmarks.forEach((b) => {
          if (b.seq > oldSeq && b.seq <= newSeq) b.seq--;
        });
      $bookmarks.forEach((b) => {
        if (b.id === current) b.seq = newSeq;
      });
    } else await fire("Error", "Failed to reorder.", "error");
  };

  const formatURL = (isSmall: boolean) => {
    const urls = Array.from(
      document.querySelectorAll(".url")
    ) as HTMLAnchorElement[];
    if (isSmall)
      urls.forEach(
        (url) => (url.text = url.text.replace(/https?:\/\/(www\.)?/i, ""))
      );
    else urls.forEach((url) => (url.text = url.dataset.url as string));
  };

  const editCategory = async (c: string) => {
    c = c.trim();
    if ($category.category != c) {
      loading.start();
      const resp = await post("/category/edit/" + $category.id, {
        category: c,
      });
      loading.end();
      let json: any = {};
      if (resp.ok) {
        json = await resp.json();
        if (json.status) {
          $bookmarks.forEach((bookmark) => {
            if (bookmark.category === $category.category) bookmark.category = c;
          });
          const index = $categories.findIndex((c) => c.id === $category.id);
          $categories[index].category = c;
          currentBookmarks = currentBookmarks;
          return true;
        }
      }
      await fire("Error", json.message ? json.message : "Error", "error");
      dispatch("reload");
      return false;
    }
    return true;
  };
  const add = () => {
    if ($category.id <= 0) $bookmark = {} as Bookmark;
    else $bookmark = { category: $category.category } as Bookmark;
    window.history.pushState({}, "", "/bookmark/add");
    $component = "bookmark";
  };
  const edit = (b: Bookmark) => {
    $bookmark = b;
    window.history.pushState({}, "", "/bookmark/edit");
    $component = "bookmark";
  };

  const categoryKeydown = async (event: KeyboardEvent) => {
    const target = event.target as Element;
    target.textContent = (target.textContent as string).trim();
    if (event.key == "Enter") {
      event.preventDefault();
      if (target.textContent)
        editable = !(await editCategory(target.textContent));
      else {
        target.textContent = $category.category;
        editable = false;
      }
    } else if (event.key == "Escape") {
      if (target.textContent) target.textContent = "";
      else {
        target.textContent = $category.category;
        editable = false;
      }
    }
  };
  const categoryClick = async () => {
    if (editable) {
      if (await confirm("category")) {
        loading.start();
        const resp = await post("/category/delete/" + $category.id);
        loading.end();
        if (resp.ok) {
          const index = $categories.findIndex((c) => c.id === $category.id);
          $categories.splice(index, 1);
          $bookmarks.forEach((bookmark) => {
            if (bookmark.category === $category.category)
              bookmark.category = "";
          });
          $bookmarks = $bookmarks;
          $categories = $categories;
          editable = false;
        } else {
          await fire("Error", await resp.text(), "error");
          dispatch("reload");
        }
        category.reset();
      }
    } else {
      editable = true;
      const target = document.querySelector("#category") as HTMLElement;
      target.setAttribute("contenteditable", "true");
      target.focus();
      const range = document.createRange();
      range.selectNodeContents(target);
      range.collapse(false);
      const sel = window.getSelection() as Selection;
      sel.removeAllRanges();
      sel.addRange(range);
    }
  };

  const handleResize = () => {
    if (smallSize != window.innerWidth <= isSmall) {
      smallSize = window.innerWidth <= isSmall;
      formatURL(smallSize);
    }
  };
  const handleScroll = async () => {
    const table = document.querySelector(".table-responsive") as Element;
    if (table.scrollTop + table.clientHeight >= table.scrollHeight)
      await bookmarks.more();
  };
  const handleClick = async (event: MouseEvent) => {
    const target = event.target as Element;
    if (
      target.id !== "category" &&
      !target.classList.contains("edit") &&
      !target.classList.contains("swal2-confirm") &&
      editable
    ) {
      const element = document.querySelector("#category") as Element;
      element.textContent = (element.textContent as string).trim();
      if (element.textContent)
        editable = !(await editCategory(element.textContent));
      else {
        target.textContent = $category.category;
        editable = false;
      }
    }
  };
</script>

<svelte:head>
  <title>
    {$category.category ? $category.category : "Uncategorized"} - My Bookmarks
  </title>
</svelte:head>

<svelte:window
  on:resize={handleResize}
  on:scroll|capture={handleScroll}
  on:click={handleClick}
/>

<div style="height: 100%">
  <header style="padding-left: 20px; height: 100px;">
    <div style="height: 50px">
      <h3
        id="category"
        class:editable
        contenteditable={editable}
        on:keydown={categoryKeydown}
      >
        {$category.category ? $category.category : "Uncategorized"}
      </h3>
      {#if $category.id > 0}
        <span class="icon" on:click={categoryClick}>
          {#if !editable}
            <i class="material-icons edit">edit</i>
          {:else}
            <i class="material-icons edit">delete</i>
          {/if}
        </span>
      {/if}
    </div>
    <button class="btn btn-primary" on:click={add}>Add Bookmark</button>
  </header>
  <div class="table-responsive">
    <table class="table table-sm">
      <thead>
        <tr>
          <th>Bookmark</th>
          <th>URL</th>
          <th>Category</th>
          <th />
        </tr>
      </thead>
      <tbody id="mybookmarks">
        {#each currentBookmarks as bookmark (bookmark.id)}
          <tr>
            <td>{bookmark.bookmark}</td>
            <td>
              <a
                href={bookmark.url}
                target="_blank"
                class="url"
                data-url={bookmark.url}>
                {bookmark.url}
              </a>
            </td>
            <td>{bookmark.category}</td>
            <td>
              <span class="icon" on:click={() => edit(bookmark)}>
                <i class="material-icons edit">edit</i>
              </span>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  </div>
</div>

<style>
  .icon {
    color: #007bff !important;
    cursor: pointer;
  }

  .icon:hover {
    color: #0056b3 !important;
  }

  .edit {
    font-size: 18px;
  }

  #category {
    outline: 0;
    display: inline-block;
    min-width: 10px;
    padding-right: 1rem;
  }

  [contenteditable="true"] {
    cursor: text;
  }

  .table-responsive {
    height: calc(100% - 100px);
    padding: 0 10px;
    cursor: default;
  }

  table {
    table-layout: fixed;
  }

  th {
    position: sticky;
    top: 0;
    border-top: 0 !important;
    border-bottom: 1px solid #dee2e6 !important;
    background-color: white;
  }

  th:nth-of-type(1) {
    width: 200px;
  }
  th:nth-of-type(3) {
    width: 200px;
  }
  th:nth-of-type(4) {
    width: 80px;
  }

  td {
    white-space: nowrap;
    text-overflow: ellipsis;
    overflow: hidden;
  }

  :global(.sortable-ghost) {
    opacity: 0;
  }

  @media (max-width: 700px) {
    th:nth-of-type(1) {
      width: 120px;
    }
    th:nth-of-type(3) {
      width: 120px;
    }
    th:nth-of-type(4) {
      width: 40px;
    }
  }
</style>
