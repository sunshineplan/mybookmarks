<script lang="ts">
  import { pasteText } from "../misc";
  import { component, showSidebar } from "../stores";
  import { category, categories } from "../bookmark";

  let hover = $state(false);

  let uncategorized = $derived($categories.find((i) => i.category === ""));
  let total = $derived($categories.reduce((a, b) => a + (b.count || 0), 0));

  const goto = async (c: Category) => {
    showSidebar.close();
    $category = c;
    window.history.pushState({}, "", "/");
    $component = "show";
    const div = document.querySelector(".table-responsive");
    if (div) div.scrollTop = 0;
  };

  const add = async (category: string) => {
    category = category.trim();
    document.querySelector(".new")?.remove();
    const newCategory: Category = { category, count: 0 };
    await categories.add(newCategory);
    await goto(newCategory);
  };

  const addCategory = async () => {
    showSidebar.close();
    const newCategory = document.querySelector<HTMLElement>(".new");
    if (newCategory) await add(newCategory.innerText);
    const ul = document.querySelector("ul.navbar-nav");
    const li = document.createElement("li");
    li.classList.add("nav-link", "new");
    const uncategorized = ul?.querySelector("#uncategorized");
    if (uncategorized) ul?.insertBefore(li, uncategorized);
    else ul?.appendChild(li);
    li.addEventListener("paste", pasteText);
    li.addEventListener("keydown", async (event) => {
      const target = event.target as Element;
      const category = target.textContent?.trim();
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
    const sel = window.getSelection();
    sel?.removeAllRanges();
    sel?.addRange(range);
  };

  const handleKeydown = async (event: KeyboardEvent) => {
    if (event.key == "ArrowUp" || event.key == "ArrowDown") {
      const newCategory = document.querySelector(".new");
      if (newCategory) newCategory.remove();
      const len = $categories.length;
      const index = $categories.findIndex(
        (c) => c.category === $category.category,
      );
      if ($component === "show")
        if (event.key == "ArrowUp") {
          if (index == 0) await goto({});
          else if (index > 0) await goto($categories[index - 1]);
        } else if (event.key == "ArrowDown")
          if ($category.category === undefined) {
            if (len > 0) await goto($categories[0]);
            else if (uncategorized) await goto({ category: "" });
          } else if (index < len - 1) await goto($categories[index + 1]);
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
        const category = newCategory.textContent?.trim();
        if (category) await add(category);
        else newCategory.remove();
      }
    }
  };
</script>

<svelte:window onkeydown={handleKeydown} onclick={handleClick} />

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<span
  class="toggle"
  onclick={showSidebar.toggle}
  onmouseenter={() => (hover = true)}
  onmouseleave={() => (hover = false)}
>
  <svg viewBox="0 0 70 70" width="40" height="30">
    {#each [10, 30, 50] as y}
      <rect {y} width="100%" height="10" fill={hover ? "#1a73e8" : "white"} />
    {/each}
  </svg>
</span>
<nav class="nav flex-column navbar-light sidebar" class:show={$showSidebar}>
  <div class="category-menu">
    <button class="btn btn-primary btn-sm" onclick={addCategory}>
      Add Category
    </button>
    <ul class="navbar-nav" id="categories">
      <!-- svelte-ignore a11y_click_events_have_key_events -->
      <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
      <li
        class="navbar-brand category"
        class:active={$category.category === undefined && $component === "show"}
        onclick={async () => await goto({})}
      >
        All Bookmarks ({total})
      </li>
      {#each $categories as c (c.category)}
        {#if c.category != ""}
          <!-- svelte-ignore a11y_click_events_have_key_events -->
          <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
          <li
            class="nav-link category"
            class:active={$category.category === c.category &&
              $component === "show"}
            onclick={async () => await goto(c)}
          >
            {c.category} ({c.count})
          </li>
        {/if}
      {/each}
      {#if uncategorized}
        <!-- svelte-ignore a11y_click_events_have_key_events -->
        <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
        <li
          class="nav-link category"
          id="uncategorized"
          class:active={$category.category === "" && $component === "show"}
          onclick={async () => await goto({ category: "" })}
        >
          Uncategorized ({uncategorized.count})
        </li>
      {/if}
    </ul>
  </div>
</nav>

<style>
  .toggle {
    display: none;
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
    --bs-navbar-brand-font-size: 1.25rem;
    --bs-navbar-brand-padding-y: 0.3125rem;
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

  @media (max-width: 900px) {
    .toggle {
      display: block;
    }

    .sidebar {
      left: -100%;
      transition: left 0.3s ease-in-out;
      box-shadow: 0 4px 8px rgba(0, 0, 0, 0.1);
    }

    .show {
      left: 0;
    }
  }
</style>
