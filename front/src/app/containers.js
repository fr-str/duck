import { createSlice } from '@reduxjs/toolkit'
export const containerSlice = createSlice({
    name: 'containers',
    initialState: {
        value: new Map(),
    },
    reducers: {
        setContainer: (state, data) => {
            state.value.set(data.payload.key, data.payload.value)
        },
        remContainer: (state, data) => {
            state.value.delete(data.payload)
        },
    },
})

// Action creators are generated for each case reducer function
export const { setContainer, remContainer } = containerSlice.actions

export const inspect = createSlice({
    name: 'inspect',
    initialState: {
        value:null,
    },
    reducers: {
        setInspect: (state, data) => {
            state.value= data.payload
        },
        clearInspect: (state, data) => {
            state.value= null
        },
    },
})

// Action creators are generated for each case reducer function
export const { setInspect, clearInspect } = inspect.actions
