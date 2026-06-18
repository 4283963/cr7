<template>
  <div class="app-container">
    <header class="app-header">
      <div class="logo">
        <span class="logo-icon">🎭</span>
        <span class="logo-text">非遗皮影戏数字排练系统</span>
      </div>
      <nav class="nav-links">
        <router-link to="/" class="nav-link">剧本列表</router-link>
      </nav>
    </header>
    <main class="app-main">
      <router-view />
    </main>
  </div>
</template>

<script setup>
</script>

<style scoped>
.app-container {
  display: flex;
  flex-direction: column;
  height: 100%;
  width: 100%;
}

.app-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  height: 60px;
  background: linear-gradient(135deg, #2d1b4e 0%, #1a1a2e 100%);
  border-bottom: 1px solid #3d2d5e;
  flex-shrink: 0;
}

.logo {
  display: flex;
  align-items: center;
  gap: 10px;
}

.logo-icon {
  font-size: 28px;
}

.logo-text {
  font-size: 18px;
  font-weight: 600;
  background: linear-gradient(90deg, #e8b059, #f4d03f);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
}

.nav-links {
  display: flex;
  gap: 8px;
}

.nav-link {
  padding: 8px 16px;
  border-radius: 6px;
  color: #b0b0c0;
  text-decoration: none;
  font-size: 14px;
  transition: all 0.2s;
}

.nav-link:hover,
.nav-link.router-link-active {
  background: rgba(232, 176, 89, 0.15);
  color: #e8b059;
}

.app-main {
  flex: 1;
  overflow: hidden;
}
</style>
