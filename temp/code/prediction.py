import pandas as pd
import xgboost as xgb
from datetime import timedelta, date
from flask import Flask, request, jsonify

app = Flask(__name__)

FEATURES = ['hour', 'min', 'rolling_mean_7_days', '6_hrs_mean', '12_hrs_mean',
       '24_hrs_mean', 'rolling_std_7d', '6_hrs_std', '12_hrs_std',
       '24_hrs_std', '24_hrs_max', '5_min_lag', '10_min_lag', '30_min_lag']
TARGET = 'Value'

reg = None
df = None
user_prediction_window = None

def create_features(df):
    df["hour"] = df.index.hour 
    df["min"] = df.index.minute
    df["rolling_mean_7_days"] = df.groupby('hour')['Value'].transform(lambda x: x.rolling(7).mean())
    df['6_hrs_mean'] = df['Value'].rolling(window=6).mean()
    df['12_hrs_mean'] = df['Value'].rolling(window=12).mean()
    df['24_hrs_mean'] = df['Value'].rolling(window=24).mean()
    df["rolling_std_7d"] = df['Value'].rolling(window=168).std()
    df['6_hrs_std'] = df['Value'].rolling(window=6).std()
    df['12_hrs_std'] = df['Value'].rolling(window=12).std()
    df['24_hrs_std'] = df['Value'].rolling(window=24).std()
    df['24_hrs_max'] = df['Value'].rolling(window=24).max()
    df['24_hrs_min'] = df['Value'].rolling(window=24).min()
    return df

def add_lags(df):
    df['5_min_lag'] = df['Value'].shift(1)
    df['10_min_lag'] = df['Value'].shift(2)
    df['30_min_lag'] = df['Value'].shift(6)
    df['5_min_lag'] = df['5_min_lag'].astype(float)
    df['10_min_lag'] = df['10_min_lag'].astype(float)
    df['30_min_lag'] = df['30_min_lag'].astype(float)
    return df


def create_model(df):
    end_date = date.today() - timedelta(days=2)
    end_date = end_date.strftime('%d/%m/%Y' )

    train = df.loc[df.index < end_date]
    test = df.loc[df.index >= end_date]


    X_train = train[FEATURES]
    Y_train = train[TARGET]  

    X_test = test[FEATURES]
    Y_test = test[TARGET]

    reg = xgb.XGBRegressor(n_estimators=1000, learning_rate=0.003)
    eval_set = [(X_test, Y_test)]
    reg.fit(X_train, Y_train, eval_metric='rmse', eval_set=eval_set, verbose=100)

    return reg

def parse_duration(duration_str):
    parts = duration_str.split("m")
    if len(parts) == 2:
        try:
            return int(parts[0])
        except ValueError:
            pass
    return 5

def get_predictions(df, reg, user_prediction_window=5):

    last_timestamp = df.index.max()
    if not isinstance(user_prediction_window, int):
        user_prediction_window = parse_duration(user_prediction_window)
    user_prediction_window = user_prediction_window//5
    for i in range(1, user_prediction_window):
        prediction_time = last_timestamp + i * pd.Timedelta(minutes=5)
        new_row = pd.DataFrame(index=[prediction_time], columns=df.columns)
        last_known_values = df.iloc[-1]
        X_to_predict = last_known_values[FEATURES].values.reshape(1, -1)
        predicted_value = reg.predict(X_to_predict)
        new_row[TARGET] = predicted_value[0]
        df = pd.concat([df, new_row])
        df = create_features(df)
        df = add_lags(df)
    return df.iloc[-1]

@app.route('/predict', methods=['POST'])
def predict():
    global df  
    try:
        data = request.get_json()

        user_prediction_window = request.headers.get('Prediction-Window')

        columns = ['Timestamp', 'Value']
        df = pd.DataFrame(data, columns=columns)

        df['Timestamp'] = pd.to_datetime(df['Timestamp'])
        df = df.set_index('Timestamp')
        
        df = create_features(df)
        df = add_lags(df)
        
        reg = create_model(df)

        prediction = get_predictions(df, reg, user_prediction_window)

        return jsonify({f'prediction for {user_prediction_window}': prediction['Value']})
    except Exception as e:
        return jsonify({'error': str(e)})
    
if __name__ == '__main__':
    app.run(host='0.0.0.0', debug=True)
