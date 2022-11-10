import { configureStore } from '@reduxjs/toolkit'
import { logSlice } from './logs'
import { containerSlice } from './containers'
import { includeContainers } from './ws'

export default configureStore({
  middleware:getDefaultMiddleware =>
  getDefaultMiddleware({
    serializableCheck: false,
  }),
  reducer: {
    logs: logSlice.reducer,
    containers: containerSlice.reducer,
    includeContainers: includeContainers.reducer,
  },
})