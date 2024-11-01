<script lang="ts">
  import Cookies from "js-cookie";
  import Sortable, { type SortableEvent } from "sortablejs";
  import { onMount } from "svelte";
  import { init, mybookmarks } from "../bookmark.svelte";
  import { confirm, loading, pasteText, poll } from "../misc.svelte";

  let { reload }: { reload: () => Promise<void> } = $props();

  const isSmall = 700;

  let smallSize = window.innerWidth <= isSmall;
  let editable = $state(false);
  let category: HTMLElement;
  let table: HTMLElement;
  let tbody: HTMLElement;

  $effect(() => {
    mybookmarks.getBookmarks(mybookmarks.category);
    table.scrollTop = 0;
    if (editable) category.focus();
  });

  onMount(() => {
    const sortable = new Sortable(tbody, {
      animation: 150,
      delay: 500,
      swapThreshold: 0.5,
      onUpdate,
    });
    if (smallSize) formatURL(true);
    return () => sortable.destroy();
  });

  const subscribe = async (signal: AbortSignal) => {
    const resp = await poll(signal);
    if (resp.ok) {
      const last = await resp.text();
      if (last && Cookies.get("last") != last) {
        const c = mybookmarks.category;
        loading.start();
        await init();
        mybookmarks.category = c;
        loading.end();
      }
      await subscribe(signal);
    } else if (resp.status == 401) {
      await reload();
    } else {
      await new Promise((sleep) => setTimeout(sleep, 30000));
      await subscribe(signal);
    }
  };
  onMount(() => {
    mybookmarks.controller = new AbortController();
    subscribe(mybookmarks.controller.signal);
    return () => mybookmarks.controller.abort();
  });

  const onUpdate = async (evt: SortableEvent) => {
    if (evt.oldIndex !== undefined && evt.newIndex !== undefined) {
      mybookmarks.controller.abort();
      await mybookmarks.swap(
        mybookmarks.bookmarks[evt.oldIndex],
        mybookmarks.bookmarks[evt.newIndex],
      );
      mybookmarks.controller = new AbortController();
      subscribe(mybookmarks.controller.signal);
    }
  };

  const formatURL = (isSmall: boolean) => {
    const urls = Array.from(
      document.querySelectorAll<HTMLAnchorElement>(".url"),
    );
    if (isSmall)
      urls.forEach(
        (url) => (url.text = url.text.replace(/https?:\/\/(www\.)?/i, "")),
      );
    else urls.forEach((url) => (url.text = url.dataset.url || ""));
  };

  const editCategory = async (c: string) => {
    c = c.trim();
    if (mybookmarks.category.category != c) {
      try {
        await mybookmarks.editCategory(mybookmarks.category, c);
        await mybookmarks.getBookmarks({ category: c });
      } catch {
        await reload();
        return false;
      }
      mybookmarks.category.category = c;
    }
    return true;
  };
  const add = () => {
    if (!mybookmarks.category.category) mybookmarks.bookmark = {} as Bookmark;
    else
      mybookmarks.bookmark = {
        category: mybookmarks.category.category,
      } as Bookmark;
    window.history.pushState({}, "", "/bookmark/add");
    mybookmarks.component = "bookmark";
  };
  const edit = (b: Bookmark) => {
    mybookmarks.bookmark = b;
    window.history.pushState({}, "", "/bookmark/edit");
    mybookmarks.component = "bookmark";
  };

  const categoryKeydown = async (event: KeyboardEvent) => {
    category.textContent = category.textContent?.trim() || "";
    if (event.key == "Enter") {
      event.preventDefault();
      if (category.textContent)
        editable = !(await editCategory(category.textContent));
      else {
        category.textContent = mybookmarks.category.category || "";
        editable = false;
      }
    } else if (event.key == "Escape") {
      if (category.textContent) category.textContent = "";
      else {
        category.textContent = mybookmarks.category.category || "";
        editable = false;
      }
    }
  };
  const categoryClick = async () => {
    if (editable) {
      if (await confirm("category")) {
        try {
          await mybookmarks.deleteCategory(mybookmarks.category);
          await mybookmarks.getBookmarks();
          editable = false;
        } catch {
          await reload();
        }
        mybookmarks.category = {};
      }
    } else {
      editable = true;
      const range = document.createRange();
      range.selectNodeContents(category);
      range.collapse(false);
      const sel = window.getSelection();
      sel?.removeAllRanges();
      sel?.addRange(range);
    }
  };

  const handleResize = () => {
    if (smallSize != window.innerWidth <= isSmall) {
      smallSize = window.innerWidth <= isSmall;
      formatURL(smallSize);
    }
  };
  const handleScroll = async () => {
    if (table.scrollTop + table.clientHeight >= table.scrollHeight)
      await mybookmarks.getBookmarks(mybookmarks.category, 15);
  };
  const handleClick = async (event: MouseEvent) => {
    const target = event.target as Element;
    if (target.classList.contains("category")) {
      editable = false;
    } else if (target.classList.contains("delete")) {
      category.textContent = mybookmarks.category.category || "";
      editable = false;
    } else if (
      !category.contains(target) &&
      !target.classList.contains("edit") &&
      editable
    ) {
      category.textContent = category.textContent?.trim() || "";
      if (category.textContent)
        editable = !(await editCategory(category.textContent));
      else {
        category.textContent = mybookmarks.category.category || "";
        editable = false;
      }
    }
  };
</script>

<svelte:head>
  <title>
    {mybookmarks.category.category === undefined
      ? "All Bookmarks"
      : mybookmarks.category.category
        ? mybookmarks.category.category
        : "Uncategorized"} - My Bookmarks
  </title>
</svelte:head>

<svelte:window
  onresize={handleResize}
  onscrollcapture={handleScroll}
  onclick={handleClick}
/>

<div style="height: 100%">
  <header style="padding-left: 20px; height: 100px;">
    <div style="height: 50px">
      <h3
        id="category"
        class:editable
        bind:this={category}
        contenteditable={editable}
        onkeydown={categoryKeydown}
        onpaste={pasteText}
      >
        {mybookmarks.category.category === undefined
          ? "All Bookmarks"
          : mybookmarks.category.category
            ? mybookmarks.category.category
            : "Uncategorized"}
      </h3>
      {#if mybookmarks.category.category}
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <span class="icon" onclick={categoryClick}>
          {#if !editable}
            <i class="material-icons edit">edit</i>
          {:else}
            <i class="material-icons delete">delete</i>
          {/if}
        </span>
      {/if}
    </div>
    <button class="btn btn-primary" onclick={add}>Add Bookmark</button>
  </header>
  <div class="table-responsive" bind:this={table}>
    <table class="table table-sm">
      <thead>
        <tr>
          <th>Bookmark</th>
          <th>URL</th>
          <th>Category</th>
          <th></th>
        </tr>
      </thead>
      <tbody bind:this={tbody}>
        {#each mybookmarks.bookmarks as bookmark (bookmark.id)}
          <tr>
            <td>{bookmark.bookmark}</td>
            <td>
              <a
                href={bookmark.url}
                target="_blank"
                rel="noreferrer"
                class="url"
                data-url={bookmark.url}
              >
                {bookmark.url}
              </a>
            </td>
            <td>{bookmark.category}</td>
            <td>
              <!-- svelte-ignore a11y_click_events_have_key_events -->
              <!-- svelte-ignore a11y_no_static_element_interactions -->
              <span class="icon" onclick={() => edit(bookmark)}>
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

  .edit,
  .delete {
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

  tbody {
    border-width: 0 !important;
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
