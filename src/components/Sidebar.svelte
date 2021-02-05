<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { fire, post } from "../misc";
  import {
    total,
    category,
    component,
    showSidebar,
    loading,
    categories,
    bookmarks,
  } from "../stores";
  import type { Category } from "../stores";

  const dispatch = createEventDispatcher();
  const isSmall = 900;

  let hover = false;
  let smallSize = window.innerWidth <= isSmall;

  $: uncategorized = $total - $categories.reduce((a, b) => a + b.count, 0);

  const goto = async (c: Category) => {
    if (window.innerWidth <= isSmall) showSidebar.close();
    $category = c;
    window.history.pushState({}, "", "/");
    $component = "show";
    await bookmarks.more(true);
  };

  const add = async (category: string) => {
    category = category.trim();
    loading.start();
    const resp = await post("/category/add", { category });
    loading.end();
    let json: any = {};
    if (resp.ok) {
      json = await resp.json();
      if (json.id && json.status) {
        (document.querySelector(".new") as Element).remove();
        const newCategory: Category = {
          id: json.id,
          category,
          count: 0,
        };
        $categories = [...$categories, newCategory];
        await goto(newCategory);
        return true;
      }
    }
    await fire("Error", json.message ? json.message : "Error", "error");
    dispatch("reload");
    return false;
  };

  const addCategory = async () => {
    if (window.innerWidth <= isSmall) showSidebar.close();
    const newCategory = document.querySelector(".new");
    let ok = true;
    if (newCategory) ok = await add((newCategory as HTMLElement).innerText);
    if (ok) {
      const ul = document.querySelector("ul.navbar-nav") as Element;
      const li = document.createElement("li");
      li.classList.add("nav-link", "new");
      const uncategorized = ul.querySelector("#uncategorized");
      if (uncategorized) ul.insertBefore(li, uncategorized);
      else ul.appendChild(li);
      li.addEventListener("keydown", async (event) => {
        const target = event.target as Element;
        const category = (target.textContent as string).trim();
        if (event.key == "Enter") {
          event.preventDefault();
          if (category) await add(category);
          else target.remove();
        } else if (event.key == "Escape") {
          if (category) target.textContent = "";
          else target.remove();
        }
      });
      li.setAttribute("contenteditable", "true");
      li.focus();
      const range = document.createRange();
      range.selectNodeContents(li);
      range.collapse(false);
      const sel = window.getSelection() as Selection;
      sel.removeAllRanges();
      sel.addRange(range);
    }
  };

  const handleKeydown = async (event: KeyboardEvent) => {
    if (event.key == "ArrowUp" || event.key == "ArrowDown") {
      const newCategory = document.querySelector(".new");
      if (newCategory) newCategory.remove();
      const len = $categories.length;
      const index = $categories.findIndex((c) => c.id === $category.id);
      if ($category.id && $component === "show")
        if (event.key == "ArrowUp") {
          if (index > 0) await goto($categories[index - 1]);
        } else if (event.key == "ArrowDown")
          if (index < len - 1) await goto($categories[index + 1]);
    }
  };
  const handleClick = async (event: MouseEvent) => {
    const target = event.target as Element;
    if (
      !target.classList.contains("new") &&
      !target.classList.contains("swal2-confirm") &&
      target.textContent !== "Add Category"
    ) {
      const newCategory = document.querySelector(".new");
      if (newCategory) {
        const category = (newCategory.textContent as string).trim();
        if (category) await add(category);
        else newCategory.remove();
      }
    }
  };
  const handleResize = () => {
    if (smallSize != window.innerWidth <= isSmall)
      smallSize = window.innerWidth <= isSmall;
  };
</script>

<svelte:window
  on:keydown={handleKeydown}
  on:click={handleClick}
  on:resize={handleResize}
/>

{#if smallSize}
  <span
    class="toggle"
    on:click={showSidebar.toggle}
    on:mouseenter={() => (hover = true)}
    on:mouseleave={() => (hover = false)}
  >
    <svg viewBox="0 0 70 70" width="40" height="30">
      {#each [10, 30, 50] as y}
        <rect {y} width="100%" height="10" fill={hover ? "#1a73e8" : "white"} />
      {/each}
    </svg>
  </span>
{/if}
<nav
  class="nav flex-column navbar-light sidebar"
  hidden={!$showSidebar && smallSize}
>
  <div class="category-menu">
    <button class="btn btn-primary btn-sm" on:click={addCategory}>
      Add Category
    </button>
    <ul class="navbar-nav" id="categories">
      <li
        class="navbar-brand category"
        class:active={$category.id === -1 && $component === "show"}
        on:click={async () =>
          await goto({ id: -1, category: "All Bookmarks", count: 0 })}
      >
        All Bookmarks ({$total})
      </li>
      {#each $categories as c (c.id)}
        <li
          class="nav-link category"
          class:active={$category.id === c.id && $component === "show"}
          on:click={async () => await goto(c)}
        >
          {c.category} ({c.count})
        </li>
      {/each}
      {#if $bookmarks.filter((b) => b.category == "").length}
        <li
          class="nav-link category"
          id="uncategorized"
          class:active={$category.id === 0 && $component === "show"}
          on:click={async () => await goto({ id: 0, category: "", count: 0 })}
        >
          Uncategorized ({uncategorized})
        </li>
      {/if}
    </ul>
  </div>
</nav>

<style>
  .toggle {
    position: fixed;
    z-index: 100;
    top: 0;
    padding: 20px;
    color: white !important;
  }

  .toggle:hover {
    background-color: rgb(232, 232, 232);
  }

  .sidebar {
    position: fixed;
    top: 0;
    z-index: 1;
    height: 100%;
    width: 250px;
    padding-top: 70px;
    user-select: none;
  }

  .category-menu {
    height: 100%;
    width: 100%;
    padding-top: 10px;
    overflow-x: hidden;
    border-right: 1px solid #e9ecef;
    background-color: white;
  }

  .category-menu .btn {
    margin-left: 20px;
    margin-bottom: 5px;
  }

  .category-menu .navbar-brand {
    text-indent: 10px;
  }

  .category-menu .navbar-nav {
    text-indent: 20px;
  }

  .category-menu .nav-link:hover {
    background-color: rgb(232, 232, 232);
  }

  #categories {
    height: calc(100% - 36px);
    overflow-y: auto;
  }

  .category {
    display: block;
    cursor: pointer;
    margin: 0;
    border-left: 5px solid transparent;
    color: rgba(0, 0, 0, 0.7) !important;
  }

  .active {
    border-left: 5px solid #1a73e8;
    color: #1a73e8 !important;
  }

  .nav-link.active {
    background-color: #eaf5fd;
  }

  :global(.new) {
    outline: 0;
    border-left: 5px solid transparent;
    background-color: #eaf5fd;
  }

  @media (min-width: 901px) {
    .sidebar {
      display: block !important;
    }
  }

  @media (max-width: 900px) {
    .sidebar {
      box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
    }
  }
</style>
