<script lang="ts">
  import Sortable from "sortablejs";
  import { onMount } from "svelte";
  import { fire, post } from "../misc";
  import { component, bookmark, bookmarks, category } from "../stores";
  import type { Bookmark } from "../stores";

  const isSmall = 700;

  let smallSize = window.innerWidth <= isSmall;

  $: currentBookmarks =
    $category.id === -1
      ? $bookmarks
      : $bookmarks.filter(
          (bookmark) => bookmark.category === $category.category
        );
  $: smallSize && formatURL(smallSize);

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
    if ((await resp.text()) == "1") console.log("reorder");
    else await fire("Error", "Failed to reorder.", "error");
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
  const editCategory = () => {
    console.log("/category/edit");
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

  const checkSize = () => {
    if (smallSize != window.innerWidth <= isSmall)
      smallSize = window.innerWidth <= isSmall;
  };
  const checkScroll = async () => {
    const table = document.querySelector(".table-responsive") as Element;
    if (table.scrollTop + table.clientHeight >= table.scrollHeight) {
      if (($category.start as number) + 30 < ($category.count as number))
        console.log("more");
    }
  };
</script>

<svelte:window on:resize={checkSize} on:scroll={checkScroll} />

<svelte:head>
  <title>{$category.category} - My Bookmarks</title>
</svelte:head>

<div style="height: 100%">
  <header style="padding-left: 20px">
    <div style="height: 50px">
      <span class="h3">{$category.category}</span>
      {#if $category.id > 0}
        <span class="btn icon" on:click={editCategory}>
          <i class="material-icons edit">edit</i>
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

  .h3 {
    cursor: default;
  }

  .edit {
    font-size: 18px;
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
