<script setup>
import { ref } from 'vue';
import { useMessage } from 'naive-ui';
import { useRouter } from 'vue-router';
import APIService from '../services/api';
import { setAuthToken, getAuthToken } from '../stores/db';
// 修改为使用模块化store
import { usePersistenceStore } from '../stores/persistence';
import { usePlayerInfoStore } from '../stores/playerInfo';

const formRef = ref(null);
const message = useMessage();
const router = useRouter();
// 使用模块化store替代原来的usePlayerStore
const persistenceStore = usePersistenceStore();
const playerInfoStore = usePlayerInfoStore();

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
    // 使用 await 确保验证完成后再执行后续操作
    await new Promise((resolve, reject) => {
      formRef.value?.validate((errors) => {
        if (!errors) {
          resolve();
        } else {
          reject(errors);
        }
      });
    });
    
    const response = await APIService.login(formValue.value.username, formValue.value.password);
    
    if (response.token) {
      // 保存令牌
      setAuthToken(response.token);
      
      // 初始化玩家数据
      // 由于 persistenceStore.initializePlayer() 不存在，我们将依赖 App.vue 在挂载时初始化数据
      
      
      
      message.success('登录成功');
      // 登录成功后跳转到游戏主界面
      router.push('/home');
    } else {
      // 明确处理登录失败的情况
      message.error(response.message || '登录失败');
      // 清除可能存储的无效令牌
      if (getAuthToken()) {
        setAuthToken(null);
      }
    }
  } catch (error) {
    if (error.message) {
      message.error('登录过程中发生错误: ' + error.message);
    } else {
      message.error('登录过程中发生未知错误');
    }
    // 确保在出现错误时清除令牌
    if (getAuthToken()) {
      setAuthToken(null);
    }
  } finally {
    loading.value = false;
  }
};

const handleRegister = async (e) => {
  e.preventDefault();
  loading.value = true;
  
  try {
    // 使用 await 确保验证完成后再执行后续操作
    await new Promise((resolve, reject) => {
      formRef.value?.validate((errors) => {
        if (!errors) {
          resolve();
        } else {
          reject(errors);
        }
      });
    });
    
    const response = await APIService.register(formValue.value.username, formValue.value.password);
    
    if (response.token) {
      // 保存令牌
      setAuthToken(response.token);
      
      // 初始化玩家数据
      // 由于 persistenceStore.initializePlayer() 不存在，我们将依赖 App.vue 在挂载时初始化数据
      
      
      
      message.success('注册成功');
      // 注册成功后跳转到首页，确保能触发新手礼包等流程
      router.push('/');
    } else {
      message.error(response.message || '注册失败');
    }
  } catch (error) {
    if (error.message) {
      message.error('注册过程中发生错误: ' + error.message);
    } else {
      message.error('注册过程中发生未知错误');
    }
  } finally {
    loading.value = false;
  }
};
</script>

<template>
  <div class="login-container">
    <n-card title="修仙界登录" style="max-width: 400px; margin: 100px auto;">
      <n-form ref="formRef" :model="formValue" :rules="rules">
        <n-form-item label="用户名" path="username">
          <n-input v-model:value="formValue.username" placeholder="请输入用户名" />
        </n-form-item>
        <n-form-item label="密码" path="password">
          <n-input v-model:value="formValue.password" type="password" placeholder="请输入密码" />
        </n-form-item>
        <n-row :gutter="[0, 12]">
          <n-col :span="24">
            <n-button type="primary" block :loading="loading" @click="handleLogin">
              登录
            </n-button>
          </n-col>
          <n-col :span="24">
            <n-button block :loading="loading" @click="handleRegister">
              注册
            </n-button>
          </n-col>
        </n-row>
      </n-form>
    </n-card>
  </div>
</template>

<style scoped>
.login-container {
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
}
</style>
