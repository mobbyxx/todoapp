import http from 'k6/http';
import { check, sleep, group } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

const errorRate = new Rate('errors');
const apiResponseTime = new Trend('api_response_time');
const authTrend = new Trend('auth_response_time');
const todoTrend = new Trend('todo_response_time');
const userTrend = new Trend('user_response_time');
const connectionTrend = new Trend('connection_response_time');
const gamificationTrend = new Trend('gamification_response_time');

const requestsCounter = new Counter('total_requests');

export const options = {
  stages: [
    { duration: '2m', target: 10 },
    { duration: '5m', target: 50 },
    { duration: '5m', target: 100 },
    { duration: '5m', target: 100 },
    { duration: '2m', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<200'],
    http_req_duration: ['p(99)<500'],
    errors: ['rate<0.1'],
    api_response_time: ['p(95)<200'],
    auth_response_time: ['p(95)<200'],
    todo_response_time: ['p(95)<200'],
    user_response_time: ['p(95)<200'],
    connection_response_time: ['p(95)<200'],
    gamification_response_time: ['p(95)<200'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

function makeRequest(method, endpoint, body = null, headers = {}) {
  const url = `${BASE_URL}${endpoint}`;
  const startTime = new Date();
  
  let response;
  if (method === 'GET') {
    response = http.get(url, { headers });
  } else if (method === 'POST') {
    response = http.post(url, JSON.stringify(body), { headers });
  } else if (method === 'PUT') {
    response = http.put(url, JSON.stringify(body), { headers });
  } else if (method === 'DELETE') {
    response = http.del(url, null, { headers });
  }
  
  const duration = new Date() - startTime;
  apiResponseTime.add(duration);
  requestsCounter.add(1);
  
  return response;
}

function registerUser() {
  group('Auth - Register', () => {
    const payload = {
      email: `loadtest_${Math.random().toString(36).substring(7)}@example.com`,
      password: 'Password123!',
      display_name: `User ${Math.random().toString(36).substring(7)}`,
    };
    
    const startTime = new Date();
    const response = makeRequest('POST', '/api/v1/auth/register', payload, {
      'Content-Type': 'application/json',
    });
    const duration = new Date() - startTime;
    authTrend.add(duration);
    
    const success = check(response, {
      'register status is 201': (r) => r.status === 201,
      'register has access_token': (r) => {
        try {
          const body = JSON.parse(r.body);
          return body.access_token !== undefined;
        } catch (e) {
          return false;
        }
      },
    });
    
    errorRate.add(!success);
    
    if (success) {
      const body = JSON.parse(response.body);
      return {
        accessToken: body.access_token,
        refreshToken: body.refresh_token,
        userId: body.user.id,
      };
    }
    return null;
  });
}

function loginUser(credentials) {
  group('Auth - Login', () => {
    const startTime = new Date();
    const response = makeRequest('POST', '/api/v1/auth/login', credentials, {
      'Content-Type': 'application/json',
    });
    const duration = new Date() - startTime;
    authTrend.add(duration);
    
    const success = check(response, {
      'login status is 200': (r) => r.status === 200,
      'login has access_token': (r) => {
        try {
          const body = JSON.parse(r.body);
          return body.access_token !== undefined;
        } catch (e) {
          return false;
        }
      },
    });
    
    errorRate.add(!success);
    
    if (success) {
      const body = JSON.parse(response.body);
      return body.access_token;
    }
    return null;
  });
}

function createTodo(authToken) {
  group('Todo - Create', () => {
    const payload = {
      title: `Load Test Todo ${Math.random().toString(36).substring(7)}`,
      description: 'This is a load test todo item',
      priority: ['low', 'medium', 'high', 'urgent'][Math.floor(Math.random() * 4)],
    };
    
    const startTime = new Date();
    const response = makeRequest('POST', '/api/v1/todos', payload, {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${authToken}`,
    });
    const duration = new Date() - startTime;
    todoTrend.add(duration);
    
    const success = check(response, {
      'create todo status is 201': (r) => r.status === 201,
      'create todo returns todo object': (r) => {
        try {
          const body = JSON.parse(r.body);
          return body.todo !== undefined;
        } catch (e) {
          return false;
        }
      },
    });
    
    errorRate.add(!success);
    
    if (success) {
      const body = JSON.parse(response.body);
      return body.todo.id;
    }
    return null;
  });
}

function listTodos(authToken) {
  group('Todo - List', () => {
    const startTime = new Date();
    const response = makeRequest('GET', '/api/v1/todos?page=1&page_size=20', null, {
      'Authorization': `Bearer ${authToken}`,
    });
    const duration = new Date() - startTime;
    todoTrend.add(duration);
    
    const success = check(response, {
      'list todos status is 200': (r) => r.status === 200,
      'list todos returns array': (r) => {
        try {
          const body = JSON.parse(r.body);
          return Array.isArray(body.todos);
        } catch (e) {
          return false;
        }
      },
    });
    
    errorRate.add(!success);
  });
}

function getTodo(authToken, todoId) {
  group('Todo - Get', () => {
    const startTime = new Date();
    const response = makeRequest('GET', `/api/v1/todos/${todoId}`, null, {
      'Authorization': `Bearer ${authToken}`,
    });
    const duration = new Date() - startTime;
    todoTrend.add(duration);
    
    const success = check(response, {
      'get todo status is 200': (r) => r.status === 200,
    });
    
    errorRate.add(!success);
  });
}

function updateTodo(authToken, todoId, version) {
  group('Todo - Update', () => {
    const payload = {
      title: `Updated Todo ${Math.random().toString(36).substring(7)}`,
      version: version,
    };
    
    const startTime = new Date();
    const response = makeRequest('PUT', `/api/v1/todos/${todoId}`, payload, {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${authToken}`,
    });
    const duration = new Date() - startTime;
    todoTrend.add(duration);
    
    const success = check(response, {
      'update todo status is 200': (r) => r.status === 200,
    });
    
    errorRate.add(!success);
  });
}

function completeTodo(authToken, todoId, version) {
  group('Todo - Complete', () => {
    const payload = { version: version };
    
    const startTime = new Date();
    const response = makeRequest('POST', `/api/v1/todos/${todoId}/complete`, payload, {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${authToken}`,
    });
    const duration = new Date() - startTime;
    todoTrend.add(duration);
    
    const success = check(response, {
      'complete todo status is 200': (r) => r.status === 200,
    });
    
    errorRate.add(!success);
  });
}

function getUserProfile(authToken) {
  group('User - Get Profile', () => {
    const startTime = new Date();
    const response = makeRequest('GET', '/api/v1/users/me', null, {
      'Authorization': `Bearer ${authToken}`,
    });
    const duration = new Date() - startTime;
    userTrend.add(duration);
    
    const success = check(response, {
      'get user profile status is 200': (r) => r.status === 200,
    });
    
    errorRate.add(!success);
  });
}

function getUserStats(authToken) {
  group('Gamification - Get Stats', () => {
    const startTime = new Date();
    const response = makeRequest('GET', '/api/v1/users/me/stats', null, {
      'Authorization': `Bearer ${authToken}`,
    });
    const duration = new Date() - startTime;
    gamificationTrend.add(duration);
    
    const success = check(response, {
      'get user stats status is 200': (r) => r.status === 200,
      'get user stats returns points': (r) => {
        try {
          const body = JSON.parse(r.body);
          return body.points !== undefined;
        } catch (e) {
          return false;
        }
      },
    });
    
    errorRate.add(!success);
  });
}

function getPointsHistory(authToken) {
  group('Gamification - Get History', () => {
    const startTime = new Date();
    const response = makeRequest('GET', '/api/v1/users/me/history?limit=20', null, {
      'Authorization': `Bearer ${authToken}`,
    });
    const duration = new Date() - startTime;
    gamificationTrend.add(duration);
    
    const success = check(response, {
      'get points history status is 200': (r) => r.status === 200,
    });
    
    errorRate.add(!success);
  });
}

function listConnections(authToken) {
  group('Connection - List', () => {
    const startTime = new Date();
    const response = makeRequest('GET', '/api/v1/connections', null, {
      'Authorization': `Bearer ${authToken}`,
    });
    const duration = new Date() - startTime;
    connectionTrend.add(duration);
    
    const success = check(response, {
      'list connections status is 200': (r) => r.status === 200,
    });
    
    errorRate.add(!success);
  });
}

export default function () {
  const credentials = {
    email: `test_${__VU}@example.com`,
    password: 'Password123!',
  };
  
  let authToken = loginUser(credentials);
  
  if (!authToken) {
    const registerResult = registerUser();
    if (registerResult) {
      authToken = registerResult.accessToken;
    }
  }
  
  if (!authToken) {
    errorRate.add(1);
    sleep(1);
    return;
  }
  
  getUserProfile(authToken);
  sleep(0.5);
  
  getUserStats(authToken);
  sleep(0.5);
  
  getPointsHistory(authToken);
  sleep(0.5);
  
  const todoIds = [];
  for (let i = 0; i < 3; i++) {
    const todoId = createTodo(authToken);
    if (todoId) {
      todoIds.push(todoId);
    }
    sleep(0.3);
  }
  
  listTodos(authToken);
  sleep(0.5);
  
  todoIds.forEach((todoId, index) => {
    getTodo(authToken, todoId);
    sleep(0.2);
    
    if (index % 2 === 0) {
      updateTodo(authToken, todoId, 1);
      sleep(0.2);
    }
    
    if (index === todoIds.length - 1) {
      completeTodo(authToken, todoId, 2);
      sleep(0.2);
    }
  });
  
  listConnections(authToken);
  sleep(0.5);
  
  sleep(Math.random() * 2 + 1);
}
