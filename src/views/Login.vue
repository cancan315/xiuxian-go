<template>
  <n-card title="用户登录">
    <n-form ref="formRef" :model="formValue" :rules="rules" label-placement="left" label-width="auto">
      <n-form-item label="用户名" path="username">
        <n-input v-model:value="formValue.username" placeholder="请输入用户名" />
      </n-form-item>
      <n-form-item label="密码" path="password">
        <n-input v-model:value="formValue.password" type="password" placeholder="请输入密码" />
      </n-form-item>
      <n-form-item>
        <n-button type="primary" @click="handleLogin" :loading="loading">登录</n-button>
        <n-button @click="handleRegister" style="margin-left: 10px" :loading="loading">注册</n-button>
      </n-form-item>
    </n-form>
  </n-card>
</template>

<script setup>
import { ref } from 'vue';
import { useMessage } from 'naive-ui';
import { useRouter } from 'vue-router';
import APIService from '../services/api';
import { setAuthToken } from '../stores/db';
import { usePlayerStore } from '../stores/player';

const formRef = ref(null);
const message = useMessage();
const router = useRouter();
const playerStore = usePlayerStore();

const loading = ref(false);

const formValue = ref({
  username: '',
  password: ''
});

const rules = {
  username: {
    required: true,
    message: '请输入用户名',
    trigger: 'blur'
  },
  password: {
    required: true,
    message: '请输入密码',
    trigger: 'blur'
  }
};

const handleLogin = async (e) => {
  e.preventDefault();
  loading.value = true;
  
  try {
    await formRef.value?.validate();
    
    const response = await APIService.login(formValue.value.username, formValue.value.password);
    
    if (response.token) {
      // Save token
      setAuthToken(response.token);
      
      // Initialize player data
      await playerStore.initializePlayer();
      
      message.success('登录成功');
      // 自动刷新页面而不是跳转
      window.location.reload();
    } else {
      message.error(response.message || '登录失败');
    }
  } catch (error) {
    message.error('登录过程中发生错误: ' + error.message);
  } finally {
    loading.value = false;
  }
};

const handleRegister = async (e) => {
  e.preventDefault();
  loading.value = true;
  
  try {
    await formRef.value?.validate();
    
    const response = await APIService.register(formValue.value.username, formValue.value.password);
    
    if (response.token) {
      // Save token
      setAuthToken(response.token);
      
      // Initialize player data
      await playerStore.initializePlayer();
      
      message.success('注册成功');
      // 自动刷新页面而不是跳转
      window.location.reload();
    } else {
      message.error(response.message || '注册失败');
    }
  } catch (error) {
    message.error('注册过程中发生错误: ' + error.message);
  } finally {
    loading.value = false;
  }
};
</script>