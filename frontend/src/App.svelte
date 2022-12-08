<script lang="ts">

import Task from "./lib/Task.svelte"
import { onMount } from "svelte";

const URL:string = "http://localhost:8080/todo";

let tasks:{ id: number; title: string; completed: boolean }[] = [];

let newTaskTitle:string = "";

async function fetchingData() {
  fetch(URL)
    .then(response => response.json())
    .then(data => {
      console.log(data);
      tasks = data;
    }).catch(error => {
      console.log(error);
    });
}

onMount(() => {
  fetchingData()
});

async function handleDelete(id: number){
    await fetch(URL + "/" + id , {
      method: "DELETE"
    });
    fetchingData();
  }

async function handleSubmit(){
  const todoItem = {
    title: newTaskTitle,
    completed: false
  };

  await fetch(URL, {
    method: "POST",
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify(todoItem),
  })

  newTaskTitle = "";

  fetchingData();
}

</script>

<main>
  <h1>todo app</h1>
  <div class="tasks">
    <form name='form' on:submit|preventDefault={handleSubmit}>
      <input name='task' bind:value={newTaskTitle} class="enter" type="text" placeholder="What to be done?" />
    </form>
    {#each tasks as t }
      <Task {handleDelete} task={t} />
    {/each}
  </div>
</main>

<style> 
  main {
    display: flex;
    align-items: center;
    flex-direction: column;
  }
  h1 {
    color: #ccc;
    font-weight: 300;
    font-size: 8rem;
  }
  .tasks {
    width: 30rem;
    box-shadow: -5px 5px 10px -5px rgb(23 54 71 / 50%);
  }
  .enter {
    width: 100%;
    padding: 0.5rem;
    border: none;
    font-size: 1.5rem;
    outline: none;
    border-bottom: 3px solid #ddd;
  }
  .enter::placeholder { 
    color: #ccc;
    font-style: italic;
    opacity: 1;
  }
</style>
