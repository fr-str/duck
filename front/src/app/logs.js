import { createSlice } from '@reduxjs/toolkit'

export const logSlice = createSlice({
    name: 'logs',
    initialState: {
        value: [],
    },
    reducers: {
        addLogs: (state, data) => {
            state.value = [...data.payload,...state.value]
        },
        popLog: (state) => {
            state.value.shift()
        },
        clearLogs: (state,conts) => {
            // remove logs that are not in the list of containers
            state.value = state.value.filter(l => conts.payload.indexOf(l.container) !== -1)
        },
    },
})

// Action creators are generated for each case reducer function
export const { addLogs, popLog,clearLogs } = logSlice.actions

export default logSlice.reducer