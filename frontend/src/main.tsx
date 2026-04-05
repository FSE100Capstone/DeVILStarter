import React from 'react'
import {createRoot} from 'react-dom/client'
import {createTheme, CssBaseline, ThemeProvider} from '@mui/material'
import './style.css'
import App from './App'

const container = document.getElementById('root')

const root = createRoot(container!)

const theme = createTheme({
    typography: {
        fontFamily: '"Nunito", "Gill Sans", "Trebuchet MS", sans-serif',
    },
});

root.render(
    <React.StrictMode>
        <ThemeProvider theme={theme}>
            <CssBaseline />
            <App/>
        </ThemeProvider>
    </React.StrictMode>
)
