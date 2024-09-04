<script lang="ts">
  import Cookies from "js-cookie";
  import Sortable, { type SortableEvent } from "sortablejs";
  import { onMount, createEventDispatcher } from "svelte";
  import { poll, confirm, pasteText } from "../misc";
  import { component, loading } from "../stores";
  import { init, bookmark, bookmarks, category, categories } from "../bookmark";

  const dispatch = createEventDispatcher();
  const isSmall = 700;

  let smallSize = window.innerWidth <= isSmall;
  let editable = false;

  $: $category, bookmarks.get($category);

  onMount(() => {
    const element = document.querySelector<HTMLElement>("#mybookmarks");
    if (element) {
      const sortable = new Sortable(element, {
        animation: 150,
        delay: 500,
        swapThreshold: 0.5,
        onUpdate,
      });
      if (smallSize) formatURL(true);
      return () => sortable.destroy();
    }
  });

  const subscribe = async (signal: AbortSignal) => {
    const resp = await poll(signal);
    if (resp.ok) {
      const last = await resp.text();
      if (last && Cookies.get("last") != last) {
        const c = $category;
        loading.start();
        await init();
        $category = c;
        loading.end();
      }
      await subscribe(signal);
    } else if (resp.status == 401) {
      dispatch("reload");
    } else {
      await new Promise((sleep) => setTimeout(sleep, 30000));
      await subscribe(signal);
    }
  };
  onMount(() => {
    const controller = new AbortController();
    subscribe(controller.signal);
    return () => controller.abort();
  });

  const onUpdate = async (evt: SortableEvent) => {
    if (evt.oldIndex !== undefined && evt.newIndex !== undefined)
      await bookmarks.swap($bookmarks[evt.oldIndex], $bookmarks[evt.newIndex]);
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
    if ($category.category != c) {
      try {
        await categories.edit($category, c);
        await bookmarks.get({ category: c });
      } catch {
        dispatch("reload");
        return false;
      }
      $category.category = c;
    }
    return true;
  };
  const add = () => {
    if (!$category.category) $bookmark = <Bookmark>{};
    else $bookmark = <Bookmark>{ category: $category.category };
    window.history.pushState({}, "", "/bookmark/add");
    $component = "bookmark";
  };
  const edit = (b: Bookmark) => {
    $bookmark = b;
    window.history.pushState({}, "", "/bookmark/edit");
    $component = "bookmark";
  };

  const categoryKeydown = async (event: KeyboardEvent) => {
    const target = <Element>event.target;
    target.textContent = target.textContent?.trim() || "";
    if (event.key == "Enter") {
      event.preventDefault();
      if (target.textContent)
        editable = !(await editCategory(target.textContent));
      else {
        target.textContent = $category.category || "";
        editable = false;
      }
    } else if (event.key == "Escape") {
      if (target.textContent) target.textContent = "";
      else {
        target.textContent = $category.category || "";
        editable = false;
      }
    }
  };
  const categoryClick = async () => {
    if (editable) {
      if (await confirm("category")) {
        try {
          await categories.delete($category);
          await bookmarks.get();
          editable = false;
        } catch {
          dispatch("reload");
        }
        $category = {};
      }
    } else {
      editable = true;
      const target = document.querySelector<HTMLElement>("#category");
      if (target) {
        target.setAttribute("contenteditable", "true");
        target.focus();
        const range = document.createRange();
        range.selectNodeContents(target);
        range.collapse(false);
        const sel = window.getSelection();
        sel?.removeAllRanges();
        sel?.addRange(range);
      }
    }
  };

  const handleResize = () => {
    if (smallSize != window.innerWidth <= isSmall) {
      smallSize = window.innerWidth <= isSmall;
      formatURL(smallSize);
    }
  };
  const handleScroll = async () => {
    const table = document.querySelector(".table-responsive");
    if (table && table.scrollTop + table.clientHeight >= table.scrollHeight)
      await bookmarks.get($category, 15);
  };
  const handleClick = async (event: MouseEvent) => {
    const target = <Element>event.target;
    const element = document.querySelector("#category");
    if (target.classList.contains("category")) {
      editable = false;
    } else if (target.classList.contains("edit")) {
      if (element) element.textContent = $category.category || "";
      editable = false;
    } else if (target.id !== "category" && editable) {
      if (element) {
        element.textContent = element.textContent?.trim() || "";
        if (element.textContent)
          editable = !(await editCategory(element.textContent));
        else {
          element.textContent = $category.category || "";
          editable = false;
        }
      }
    }
  };
</script>

<svelte:head>
  <title>
    {$category.category === undefined
      ? "All Bookmarks"
      : $category.category
        ? $category.category
        : "Uncategorized"} - My Bookmarks
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
        on:paste={pasteText}
      >
        {$category.category === undefined
          ? "All Bookmarks"
          : $category.category
            ? $category.category
            : "Uncategorized"}
      </h3>
      {#if $category.category}
        <!-- svelte-ignore a11y-click-events-have-key-events -->
        <!-- svelte-ignore a11y-no-static-element-interactions -->
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
        {#each $bookmarks as bookmark (bookmark.id)}
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
              <!-- svelte-ignore a11y-click-events-have-key-events -->
              <!-- svelte-ignore a11y-no-static-element-interactions -->
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
