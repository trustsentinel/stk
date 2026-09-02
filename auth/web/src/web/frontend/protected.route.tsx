import React, { ReactNode } from 'react';
import { Route, RouteProps } from 'react-router-dom';
import auth from './auth';
import otp from './otp';

const ProtectedRoute = ({ component: Component, ...rest}: RouteProps) => {
    if (!Component) return null;
    return (
        <Route
        {...rest}
        render={ (props) => {
            if(auth.isAuthenticated()) {
                if(otp.isAuthenticated()) {

                }
                return <Component {...props} />;
            }

        }}
        />
    )
};

export default ProtectedRoute;