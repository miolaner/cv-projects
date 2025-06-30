{
    'name': 'Todo App',
    'version': '16.0.1.0.0',
    'category': 'Productivity',
    'summary': 'Manage your tasks with time tracking',
    'description': """
        Todo Application with the following features:
        - Create tasks with name, description, start time, and end time
        - Edit existing tasks
        - Delete tasks
        - Track time spent on tasks
        - User authentication
    """,
    'author': 'Your Name',
    'depends': ['base', 'mail'],
    'data': [
        'security/ir.model.access.csv',
        'views/todo_task_views.xml',
        'views/menu_views.xml',
    ],
    'application': True,
    'installable': True,
}
