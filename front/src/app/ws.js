import { createSlice } from '@reduxjs/toolkit'

export const includeContainers = createSlice({
    name: 'includeContainers',
    initialState: {
        value: [],
    },
    reducers: {
        addCont: (state, cont) => {
            // unique
            if (state.value.indexOf(cont.payload) === -1) {
                state.value.push(cont.payload)
            }
            window.localStorage.setItem('includeContainers', JSON.stringify(state.value))
        },
        remCont: (state,cont) => {
            state.value = state.value.filter(c => c !== cont.payload)
            window.localStorage.setItem('includeContainers', JSON.stringify(state.value))
        },
        setConts: (state,conts) => {
            state.value = conts.payload
        }
    },
})

// Action creators are generated for each case reducer function
export const { addCont, remCont,setConts } = includeContainers.actions

export default includeContainers.reducer