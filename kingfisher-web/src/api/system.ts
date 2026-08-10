import request from './request';

export const systemApi = {
  getInfo: () => request.get('/system/info'),
};
