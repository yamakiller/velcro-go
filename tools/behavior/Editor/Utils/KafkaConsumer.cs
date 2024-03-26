using Confluent.Kafka;
using System;
using System.Collections.Generic;
using System.Configuration;
using System.Diagnostics;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace Editor.Utils
{
    public class KafkaConsumer
    {
        public static string brokerUrl = "127.0.0.1:9092";
        public static string topic = "BehaviorNode";// ConfigurationManager.AppSettings["ConsumeTopic"];
        public static string groupid = "Behavior";//  ConfigurationManager.AppSettings["GroupID"];
        private static CancellationTokenSource? cts = null;
        public static event EventHandler<EventArgs> CallbackEvent;
        private static Task _task;
        public static async Task StartConsume(string url)
        {
            var brokerList = brokerUrl;
            if (url != "")
            {
                brokerList = url;
            }      
            List<string> topics = new List<string>(topic.Split(','));
            
            try
            {
                cts = new CancellationTokenSource();

                _task = Task.Run(() => RunConsume(brokerList, topics, cts.Token));
            }
            catch (Exception ex)
            {
                StopConsume();
            }

        }
        /// <summary>
        ///     In this example
        ///         - offsets are manually committed.
        ///         - no extra thread is created for the Poll (Consume) loop.
        /// </summary>
        private static async Task RunConsume(string brokerList, List<string> topics, CancellationToken cancellationToken)
        {
            var config = new ConsumerConfig
            {
                BootstrapServers = brokerList,
                GroupId = groupid,
                EnableAutoCommit = false,
                StatisticsIntervalMs = 5000,
                SessionTimeoutMs = 6000,
                AutoOffsetReset =AutoOffsetReset.Earliest,
            };

            using (var consumer = new ConsumerBuilder<Ignore, string>(config).SetErrorHandler((_, e) => Console.WriteLine($"Error: {e.Reason}")).Build())
            {
                consumer.Subscribe(topics);
                try
                {
                    while (!cancellationToken.IsCancellationRequested)
                    {
                        try
                        {
                            var consumeResult = consumer.Consume(cancellationToken);
                            Console.WriteLine($"Received message at {consumeResult.TopicPartitionOffset}: ${consumeResult.Message.Value}");
                            consumer.StoreOffset(consumeResult);//<----this
                            CallbackEvent?.Invoke(consumeResult.Message.Value,null);
                        }
                        catch (ConsumeException e)
                        {
                            Console.WriteLine($"Error occured: {e.Error.Reason}");
                        }
                    }
                }
                catch (OperationCanceledException)
                {
                    // Ensure the consumer leaves the group cleanly and final offsets are committed.
                    consumer.Close();
                }
            }
        }

        public static void StopConsume()
        {
            try
            {
                cts.Cancel();
                _task.Dispose();
            }
            catch (Exception)
            {

            }
        }
    }

}
